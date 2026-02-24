package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/pkg/logger"

	"github.com/google/uuid"
)

// QualityDimension 质量评分维度。
type QualityDimension struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Score      int    `json:"score"`
	MaxScore   int    `json:"max_score"`
	Comment    string `json:"comment"`
	Importance int    `json:"importance"`
}

// LessonQualityReview 教案质量审查结果。
type LessonQualityReview struct {
	LessonID     uuid.UUID          `json:"lesson_id"`
	TotalScore   int                `json:"total_score"`
	MaxScore     int                `json:"max_score"`
	Grade        string             `json:"grade"`
	Dimensions   []QualityDimension `json:"dimensions"`
	Issues       []string           `json:"issues"`
	Suggestions  []string           `json:"suggestions"`
	AutoApproved bool               `json:"auto_approved"`
}

// VersionDiffField 版本差异字段。
type VersionDiffField struct {
	Field   string   `json:"field"`
	Label   string   `json:"label"`
	Changed bool     `json:"changed"`
	Before  string   `json:"before,omitempty"`
	After   string   `json:"after,omitempty"`
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

// LessonVersionDiff 教案版本差异结果。
type LessonVersionDiff struct {
	LessonID      uuid.UUID          `json:"lesson_id"`
	FromVersion   string             `json:"from_version"`
	ToVersion     string             `json:"to_version"`
	ChangedFields int                `json:"changed_fields"`
	Fields        []VersionDiffField `json:"fields"`
}

type agentLessonQualityReviewRequest struct {
	LessonID   string `json:"lessonId"`
	Title      string `json:"title"`
	Subject    string `json:"subject"`
	Grade      string `json:"grade"`
	Duration   int    `json:"duration"`
	Objectives string `json:"objectives"`
	Content    string `json:"content"`
	Activities string `json:"activities"`
	Assessment string `json:"assessment"`
	Resources  string `json:"resources"`
}

type agentLessonQualityReviewResponse struct {
	Success bool                 `json:"success"`
	Data    *LessonQualityReview `json:"data"`
	Error   string               `json:"error,omitempty"`
}

func calculateQualityGrade(totalScore int) string {
	switch {
	case totalScore >= 85:
		return "A"
	case totalScore >= 70:
		return "B"
	case totalScore >= 55:
		return "C"
	default:
		return "D"
	}
}

func normalizeQualityReviewResult(review *LessonQualityReview) error {
	if review == nil {
		return errors.New("质量审查结果为空")
	}
	if len(review.Dimensions) == 0 {
		return errors.New("质量审查维度为空")
	}

	totalScore := 0
	maxScore := 0
	for i := range review.Dimensions {
		dimension := &review.Dimensions[i]
		if dimension.MaxScore < 0 {
			dimension.MaxScore = 0
		}
		if dimension.Score < 0 {
			dimension.Score = 0
		}
		if dimension.MaxScore > 0 && dimension.Score > dimension.MaxScore {
			dimension.Score = dimension.MaxScore
		}
		totalScore += dimension.Score
		maxScore += dimension.MaxScore
	}

	if maxScore <= 0 {
		return errors.New("质量审查最大分数无效")
	}

	review.TotalScore = totalScore
	review.MaxScore = maxScore
	review.Grade = calculateQualityGrade(totalScore)
	review.AutoApproved = totalScore >= 75

	if len(review.Issues) == 0 {
		review.Issues = []string{"未发现明显结构性问题。"}
	}
	if len(review.Suggestions) == 0 {
		if review.AutoApproved {
			review.Suggestions = []string{"整体质量较好，可进入人工抽检或直接发布。"}
		} else {
			review.Suggestions = []string{"建议根据低分维度逐项修订后再发布。"}
		}
	}

	return nil
}

func (s *lessonService) reviewQualityByAgent(ctx context.Context, lesson *model.Lesson) (*LessonQualityReview, error) {
	if s.cfg == nil || strings.TrimSpace(s.cfg.URL) == "" || s.httpClient == nil {
		return nil, errors.New("agent 评分服务未配置")
	}

	requestPayload := agentLessonQualityReviewRequest{
		LessonID:   lesson.ID.String(),
		Title:      strings.TrimSpace(lesson.Title),
		Subject:    strings.TrimSpace(lesson.Subject),
		Grade:      strings.TrimSpace(lesson.Grade),
		Duration:   lesson.Duration,
		Objectives: normalizeLessonText(lesson.Objectives),
		Content:    normalizeLessonText(lesson.Content),
		Activities: normalizeLessonText(lesson.Activities),
		Assessment: normalizeLessonText(lesson.Assessment),
		Resources:  normalizeLessonText(lesson.Resources),
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("marshal quality review request failed: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if s.cfg.APIKey != "" {
		headers["Authorization"] = "Bearer " + s.cfg.APIKey
	}

	url := fmt.Sprintf("%s/api/quality-review", strings.TrimRight(s.cfg.URL, "/"))
	statusCode, respBody, err := doAgentRequestWithRetry(
		ctx,
		s.httpClient,
		http.MethodPost,
		url,
		body,
		headers,
		"quality_review",
	)
	if err != nil {
		return nil, fmt.Errorf("call quality review endpoint failed: %w", err)
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("quality review endpoint returned error: %d - %s", statusCode, string(respBody))
	}

	var response agentLessonQualityReviewResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("unmarshal quality review response failed: %w", err)
	}
	if !response.Success {
		if strings.TrimSpace(response.Error) != "" {
			return nil, errors.New(strings.TrimSpace(response.Error))
		}
		return nil, errors.New("quality review failed")
	}
	if response.Data == nil {
		return nil, errors.New("quality review response is empty")
	}

	response.Data.LessonID = lesson.ID
	if err := normalizeQualityReviewResult(response.Data); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func normalizeLessonText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &jsonObj); err == nil {
		if textValue, ok := jsonObj["text"].(string); ok {
			return strings.TrimSpace(textValue)
		}
		encoded, _ := json.MarshalIndent(jsonObj, "", "  ")
		return strings.TrimSpace(string(encoded))
	}

	return raw
}

func containsAnyKeyword(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func splitNonEmptyLines(text string) []string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}

func truncateDiffText(text string) string {
	const maxChars = 4000
	if len(text) <= maxChars {
		return text
	}
	return text[:maxChars] + "\n...（已截断）"
}

func computeLineDelta(beforeText, afterText string) (added []string, removed []string) {
	beforeLines := splitNonEmptyLines(beforeText)
	afterLines := splitNonEmptyLines(afterText)
	if len(beforeLines) == 0 && len(afterLines) == 0 {
		return nil, nil
	}

	dp := make([][]int, len(beforeLines)+1)
	for i := range dp {
		dp[i] = make([]int, len(afterLines)+1)
	}

	for i := len(beforeLines) - 1; i >= 0; i-- {
		for j := len(afterLines) - 1; j >= 0; j-- {
			if beforeLines[i] == afterLines[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				if dp[i+1][j] > dp[i][j+1] {
					dp[i][j] = dp[i+1][j]
				} else {
					dp[i][j] = dp[i][j+1]
				}
			}
		}
	}

	i, j := 0, 0
	for i < len(beforeLines) && j < len(afterLines) {
		switch {
		case beforeLines[i] == afterLines[j]:
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			removed = append(removed, beforeLines[i])
			i++
		default:
			added = append(added, afterLines[j])
			j++
		}
	}

	for ; i < len(beforeLines); i++ {
		removed = append(removed, beforeLines[i])
	}
	for ; j < len(afterLines); j++ {
		added = append(added, afterLines[j])
	}

	if len(added) > 30 {
		added = append(added[:30], "...（新增行过多，已截断）")
	}
	if len(removed) > 30 {
		removed = append(removed[:30], "...（删除行过多，已截断）")
	}

	return added, removed
}

func (s *lessonService) ReviewQuality(ctx context.Context, lessonID uuid.UUID, userID uuid.UUID) (*LessonQualityReview, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, ErrLessonNotFound
	}
	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}

	// 优先使用 Agent 进行语义评分；失败时回退本地规则评分，避免功能不可用。
	if review, agentErr := s.reviewQualityByAgent(ctx, lesson); agentErr == nil {
		return review, nil
	} else {
		logger.Warn(
			"Agent quality review failed, fallback to rule-based scoring",
			logger.String("lesson_id", lesson.ID.String()),
			logger.Err(agentErr),
		)
	}

	objectivesText := normalizeLessonText(lesson.Objectives)
	contentText := normalizeLessonText(lesson.Content)
	activitiesText := normalizeLessonText(lesson.Activities)
	assessmentText := normalizeLessonText(lesson.Assessment)
	resourcesText := normalizeLessonText(lesson.Resources)

	dimensions := make([]QualityDimension, 0, 6)
	issues := make([]string, 0)
	suggestions := make([]string, 0)

	// 1) 完整度
	completenessScore := 0
	nonEmptyCount := 0
	fields := []string{objectivesText, contentText, activitiesText, assessmentText, resourcesText}
	for _, field := range fields {
		if strings.TrimSpace(field) != "" {
			nonEmptyCount++
		}
	}
	completenessScore = nonEmptyCount * 5
	if completenessScore > 25 {
		completenessScore = 25
	}
	if completenessScore < 20 {
		issues = append(issues, "教案关键章节不完整，部分内容缺失。")
		suggestions = append(suggestions, "补齐教学目标、教学活动、教学评价与教学资源四个核心章节。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "completeness",
		Name:       "内容完整度",
		Score:      completenessScore,
		MaxScore:   25,
		Comment:    "检查目标、内容、活动、评价、资源是否完整。",
		Importance: 5,
	})

	// 2) 目标可衡量性
	objectiveScore := 0
	objectiveLen := len([]rune(objectivesText))
	if objectiveLen > 20 {
		objectiveScore += 8
	}
	if containsAnyKeyword(objectivesText, []string{"能够", "掌握", "理解", "分析", "表达", "应用"}) {
		objectiveScore += 7
	}
	if containsAnyKeyword(objectivesText, []string{"知识", "过程", "情感", "价值观"}) {
		objectiveScore += 5
	}
	if objectiveScore > 20 {
		objectiveScore = 20
	}
	if objectiveScore < 14 {
		issues = append(issues, "教学目标描述偏弱，可测量性不足。")
		suggestions = append(suggestions, "使用“能够/掌握/完成/表达”等可观察动词重写目标。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "objective_clarity",
		Name:       "目标清晰度",
		Score:      objectiveScore,
		MaxScore:   20,
		Comment:    "目标应具体、可评估并覆盖多维目标。",
		Importance: 5,
	})

	// 3) 活动设计质量
	activityScore := 0
	activityLen := len([]rune(activitiesText))
	if activityLen > 40 {
		activityScore += 8
	}
	if containsAnyKeyword(activitiesText, []string{"讨论", "合作", "小组", "实验", "探究", "展示", "提问"}) {
		activityScore += 7
	}
	if containsAnyKeyword(contentText, []string{"导入", "讲解", "练习", "总结", "作业"}) {
		activityScore += 5
	}
	if activityScore > 20 {
		activityScore = 20
	}
	if activityScore < 14 {
		issues = append(issues, "教学活动设计偏弱，互动性或流程性不足。")
		suggestions = append(suggestions, "补充导入-探究-练习-总结链路，并增加师生互动环节。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "activity_design",
		Name:       "活动设计",
		Score:      activityScore,
		MaxScore:   20,
		Comment:    "强调教学流程完整性与课堂互动密度。",
		Importance: 4,
	})

	// 4) 评价对齐度
	assessmentScore := 0
	if len([]rune(assessmentText)) > 20 {
		assessmentScore += 7
	}
	if containsAnyKeyword(assessmentText, []string{"评价", "反馈", "检测", "形成性", "达成"}) {
		assessmentScore += 5
	}
	if containsAnyKeyword(assessmentText+objectivesText, []string{"目标", "达成"}) {
		assessmentScore += 3
	}
	if assessmentScore > 15 {
		assessmentScore = 15
	}
	if assessmentScore < 10 {
		issues = append(issues, "教学评价与目标对齐程度不高。")
		suggestions = append(suggestions, "补充形成性评价标准，并明确评价证据对应教学目标。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "assessment_alignment",
		Name:       "评价对齐度",
		Score:      assessmentScore,
		MaxScore:   15,
		Comment:    "评价应与目标和活动形成闭环。",
		Importance: 4,
	})

	// 5) 资源支撑度
	resourceScore := 0
	resourceLines := splitNonEmptyLines(resourcesText)
	if len(resourceLines) >= 1 {
		resourceScore += 5
	}
	if len(resourceLines) >= 3 {
		resourceScore += 5
	}
	if resourceScore < 6 {
		issues = append(issues, "教学资源描述较少，可能影响落地执行。")
		suggestions = append(suggestions, "补充课件、教具、练习材料及扩展阅读资源。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "resource_support",
		Name:       "资源支撑",
		Score:      resourceScore,
		MaxScore:   10,
		Comment:    "资源越具体，教学执行越稳定。",
		Importance: 3,
	})

	// 6) 课时合理性
	durationScore := 0
	switch {
	case lesson.Duration >= 35 && lesson.Duration <= 55:
		durationScore = 10
	case lesson.Duration >= 30 && lesson.Duration <= 70:
		durationScore = 7
	default:
		durationScore = 4
		issues = append(issues, "课时设置与常规课堂时长偏差较大。")
		suggestions = append(suggestions, "建议调整课时或拆分教学内容，避免课堂负荷过重。")
	}
	dimensions = append(dimensions, QualityDimension{
		Key:        "time_feasibility",
		Name:       "课时可执行性",
		Score:      durationScore,
		MaxScore:   10,
		Comment:    "时长应匹配课堂节奏与活动密度。",
		Importance: 3,
	})

	totalScore := 0
	maxScore := 0
	for _, dimension := range dimensions {
		totalScore += dimension.Score
		maxScore += dimension.MaxScore
	}

	grade := calculateQualityGrade(totalScore)

	autoApproved := totalScore >= 75
	if autoApproved {
		suggestions = append(suggestions, "整体质量较好，可进入人工抽检或直接发布。")
	} else {
		suggestions = append(suggestions, "建议根据低分维度逐项修订后再发布。")
	}

	if len(issues) == 0 {
		issues = append(issues, "未发现明显结构性问题。")
	}

	return &LessonQualityReview{
		LessonID:     lesson.ID,
		TotalScore:   totalScore,
		MaxScore:     maxScore,
		Grade:        grade,
		Dimensions:   dimensions,
		Issues:       issues,
		Suggestions:  suggestions,
		AutoApproved: autoApproved,
	}, nil
}

func parseVersionToken(token string) (int, bool, error) {
	value := strings.TrimSpace(strings.ToLower(token))
	if value == "" {
		return 0, false, errors.New("版本号不能为空")
	}
	if value == "current" || value == "latest" {
		return 0, true, nil
	}

	value = strings.TrimPrefix(value, "v")
	version, err := strconv.Atoi(value)
	if err != nil || version <= 0 {
		return 0, false, errors.New("版本号格式错误")
	}
	return version, false, nil
}

func parseLessonSnapshot(raw string) (map[string]interface{}, error) {
	var snapshot map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func normalizeSnapshotField(snapshot map[string]interface{}, key string) string {
	value, exists := snapshot[key]
	if !exists || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		if key == "objectives" || key == "content" {
			return normalizeLessonText(typed)
		}
		return strings.TrimSpace(typed)
	case float64:
		if key == "duration" {
			return strconv.Itoa(int(typed))
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	case []interface{}:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, fmt.Sprintf("%v", item))
		}
		return strings.Join(items, ", ")
	case map[string]interface{}:
		body, _ := json.MarshalIndent(typed, "", "  ")
		return string(body)
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func (s *lessonService) resolveSnapshot(ctx context.Context, lesson *model.Lesson, versionToken string) (string, map[string]interface{}, error) {
	version, isCurrent, err := parseVersionToken(versionToken)
	if err != nil {
		return "", nil, err
	}

	if isCurrent {
		raw, err := buildLessonSnapshot(lesson)
		if err != nil {
			return "", nil, err
		}
		snapshot, err := parseLessonSnapshot(raw)
		if err != nil {
			return "", nil, err
		}
		return "current", snapshot, nil
	}

	if s.versionRepo == nil {
		return "", nil, errors.New("版本功能未启用")
	}
	versionData, err := s.versionRepo.GetByVersion(ctx, lesson.ID, version)
	if err != nil {
		return "", nil, errors.New("版本不存在")
	}

	snapshot, err := parseLessonSnapshot(versionData.Content)
	if err != nil {
		return "", nil, fmt.Errorf("解析版本快照失败: %w", err)
	}

	return fmt.Sprintf("v%d", version), snapshot, nil
}

func (s *lessonService) CompareVersions(ctx context.Context, lessonID uuid.UUID, userID uuid.UUID, fromVersion, toVersion string) (*LessonVersionDiff, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, ErrLessonNotFound
	}
	if lesson.UserID != userID {
		return nil, ErrUnauthorized
	}

	fromLabel, fromSnapshot, err := s.resolveSnapshot(ctx, lesson, fromVersion)
	if err != nil {
		return nil, err
	}
	toLabel, toSnapshot, err := s.resolveSnapshot(ctx, lesson, toVersion)
	if err != nil {
		return nil, err
	}

	fieldDefs := []struct {
		key   string
		label string
	}{
		{key: "title", label: "标题"},
		{key: "subject", label: "学科"},
		{key: "grade", label: "年级"},
		{key: "duration", label: "课时"},
		{key: "status", label: "状态"},
		{key: "objectives", label: "教学目标"},
		{key: "content", label: "教学内容"},
		{key: "activities", label: "教学活动"},
		{key: "assessment", label: "教学评价"},
		{key: "resources", label: "教学资源"},
		{key: "tags", label: "标签"},
	}

	fields := make([]VersionDiffField, 0, len(fieldDefs))
	changedCount := 0

	for _, field := range fieldDefs {
		before := normalizeSnapshotField(fromSnapshot, field.key)
		after := normalizeSnapshotField(toSnapshot, field.key)
		changed := before != after
		added, removed := computeLineDelta(before, after)
		if changed {
			changedCount++
		}

		fields = append(fields, VersionDiffField{
			Field:   field.key,
			Label:   field.label,
			Changed: changed,
			Before:  truncateDiffText(before),
			After:   truncateDiffText(after),
			Added:   added,
			Removed: removed,
		})
	}

	sort.SliceStable(fields, func(i, j int) bool {
		if fields[i].Changed != fields[j].Changed {
			return fields[i].Changed
		}
		return fields[i].Label < fields[j].Label
	})

	return &LessonVersionDiff{
		LessonID:      lesson.ID,
		FromVersion:   fromLabel,
		ToVersion:     toLabel,
		ChangedFields: changedCount,
		Fields:        fields,
	}, nil
}
