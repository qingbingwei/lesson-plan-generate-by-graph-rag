package service

import "fmt"

// AgentRequest Agent请求
type AgentRequest struct {
	Subject    string   `json:"subject"`
	Grade      string   `json:"grade"`
	Topic      string   `json:"topic"`
	Duration   int      `json:"duration"`
	Objectives []string `json:"objectives"`
	Keywords   []string `json:"keywords"`
	Style      string   `json:"style"`
	Difficulty string   `json:"difficulty"`
	UserId     string   `json:"userId"`
}

// AgentResponse Agent响应
type AgentResponse struct {
	Success bool                 `json:"success"`
	Data    *GeneratedLessonData `json:"data"`
	Error   string               `json:"error,omitempty"`
	Usage   *TokenUsage          `json:"usage,omitempty"`
}

// GeneratedLessonData 生成的教案数据
type GeneratedLessonData struct {
	Title           string           `json:"title"`
	Objectives      LessonObjectives `json:"objectives"`
	KeyPoints       []string         `json:"keyPoints"`
	DifficultPoints []string         `json:"difficultPoints"`
	TeachingMethods []string         `json:"teachingMethods"`
	Content         LessonContent    `json:"content"`
	Evaluation      string           `json:"evaluation"`
	Reflection      string           `json:"reflection,omitempty"`
}

// LessonObjectives 教学目标
type LessonObjectives struct {
	Knowledge string `json:"knowledge"`
	Process   string `json:"process"`
	Emotion   string `json:"emotion"`
}

// LessonContent 教学内容
type LessonContent struct {
	Sections  []LessonSection `json:"sections"`
	Materials []string        `json:"materials"`
	Homework  string          `json:"homework"`
}

// LessonSection 教学环节
type LessonSection struct {
	Title           string `json:"title"`
	Duration        int    `json:"duration"`
	TeacherActivity string `json:"teacherActivity"`
	StudentActivity string `json:"studentActivity"`
	Content         string `json:"content"`
	DesignIntent    string `json:"designIntent,omitempty"`
}

// TokenUsage Token使用情况
type TokenUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// 辅助函数：格式化教学目标
func FormatObjectives(obj LessonObjectives) string {
	result := ""
	if obj.Knowledge != "" {
		result += "【知识与技能】\n" + obj.Knowledge + "\n\n"
	}
	if obj.Process != "" {
		result += "【过程与方法】\n" + obj.Process + "\n\n"
	}
	if obj.Emotion != "" {
		result += "【情感态度价值观】\n" + obj.Emotion
	}
	return result
}

// 辅助函数：格式化字符串列表
func FormatStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%d. %s", i+1, item)
	}
	return result
}

// 辅助函数：格式化教学环节
func FormatSections(sections []LessonSection) string {
	result := ""
	for i, section := range sections {
		if i > 0 {
			result += "\n\n"
		}
		result += fmt.Sprintf("## %s (%d分钟)\n\n", section.Title, section.Duration)
		if section.Content != "" {
			result += section.Content + "\n\n"
		}
		if section.TeacherActivity != "" {
			result += "**教师活动：**\n" + section.TeacherActivity + "\n\n"
		}
		if section.StudentActivity != "" {
			result += "**学生活动：**\n" + section.StudentActivity
		}
	}
	return result
}

// 辅助函数：格式化学生活动
func FormatActivities(sections []LessonSection) string {
	result := ""
	for i, section := range sections {
		if section.StudentActivity != "" {
			if i > 0 {
				result += "\n\n"
			}
			result += fmt.Sprintf("**%s阶段：**\n%s", section.Title, section.StudentActivity)
		}
	}
	return result
}

// 辅助函数：格式化教学材料
func FormatMaterials(materials []string) string {
	result := ""
	for i, material := range materials {
		if i > 0 {
			result += "\n"
		}
		result += fmt.Sprintf("%d. %s", i+1, material)
	}
	return result
}
