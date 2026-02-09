package repository

import (
	"context"
	"fmt"

	"lesson-plan/backend/internal/config"
	"lesson-plan/backend/internal/model"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// KnowledgeRepository 知识点仓库接口
type KnowledgeRepository interface {
	Create(ctx context.Context, knowledge *model.Knowledge) error
	GetByID(ctx context.Context, id string) (*model.Knowledge, error)
	Update(ctx context.Context, knowledge *model.Knowledge) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, limit int) ([]model.Knowledge, error)
	SearchByEmbedding(ctx context.Context, embedding []float64, limit int) ([]model.Knowledge, error)
	GetRelated(ctx context.Context, id string, limit int) ([]model.Knowledge, error)
	CreateRelation(ctx context.Context, relation *model.KnowledgeRelation) error
	GetGraph(ctx context.Context, subject, grade, userId string, limit int) (*model.KnowledgeGraph, error)
}

type knowledgeRepository struct {
	driver   neo4j.DriverWithContext
	database string
}

// NewKnowledgeRepository 创建知识点仓库
func NewKnowledgeRepository(driver neo4j.DriverWithContext, cfg *config.Neo4jConfig) KnowledgeRepository {
	return &knowledgeRepository{
		driver:   driver,
		database: cfg.Database,
	}
}

func (r *knowledgeRepository) session(ctx context.Context) neo4j.SessionWithContext {
	// 如果没有指定数据库，使用默认的 neo4j 数据库
	dbName := r.database
	if dbName == "" {
		dbName = "neo4j"
	}
	return r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: dbName})
}

func (r *knowledgeRepository) Create(ctx context.Context, knowledge *model.Knowledge) error {
	session := r.session(ctx)
	defer session.Close(ctx)

	query := `
		CREATE (k:Knowledge {
			id: $id,
			name: $name,
			type: $type,
			subject: $subject,
			grade: $grade,
			description: $description,
			keywords: $keywords,
			embedding: $embedding,
			created_at: datetime(),
			updated_at: datetime()
		})
		RETURN k
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          knowledge.ID,
			"name":        knowledge.Name,
			"type":        knowledge.Type,
			"subject":     knowledge.Subject,
			"grade":       knowledge.Grade,
			"description": knowledge.Description,
			"keywords":    knowledge.Keywords,
			"embedding":   knowledge.Embedding,
		})
		return nil, err
	})

	return err
}

func (r *knowledgeRepository) GetByID(ctx context.Context, id string) (*model.Knowledge, error) {
	session := r.session(ctx)
	defer session.Close(ctx)

	query := `MATCH (k:Knowledge {id: $id}) RETURN k`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		if err != nil {
			return nil, err
		}

		if records.Next(ctx) {
			node, _ := records.Record().Get("k")
			return r.nodeToKnowledge(node.(neo4j.Node)), nil
		}

		return nil, fmt.Errorf("knowledge not found: %s", id)
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.Knowledge), nil
}

func (r *knowledgeRepository) Update(ctx context.Context, knowledge *model.Knowledge) error {
	session := r.session(ctx)
	defer session.Close(ctx)

	query := `
		MATCH (k:Knowledge {id: $id})
		SET k.name = $name,
			k.type = $type,
			k.subject = $subject,
			k.grade = $grade,
			k.description = $description,
			k.keywords = $keywords,
			k.embedding = $embedding,
			k.updated_at = datetime()
		RETURN k
	`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          knowledge.ID,
			"name":        knowledge.Name,
			"type":        knowledge.Type,
			"subject":     knowledge.Subject,
			"grade":       knowledge.Grade,
			"description": knowledge.Description,
			"keywords":    knowledge.Keywords,
			"embedding":   knowledge.Embedding,
		})
		return nil, err
	})

	return err
}

func (r *knowledgeRepository) Delete(ctx context.Context, id string) error {
	session := r.session(ctx)
	defer session.Close(ctx)

	query := `MATCH (k:Knowledge {id: $id}) DETACH DELETE k`

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, query, map[string]interface{}{"id": id})
		return nil, err
	})

	return err
}

func (r *knowledgeRepository) Search(ctx context.Context, query string, limit int) ([]model.Knowledge, error) {
	session := r.session(ctx)
	defer session.Close(ctx)

	cypher := `
		MATCH (k:Knowledge)
		WHERE k.name CONTAINS $query OR k.description CONTAINS $query
		RETURN k
		LIMIT $limit
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, cypher, map[string]interface{}{
			"query": query,
			"limit": limit,
		})
		if err != nil {
			return nil, err
		}

		var knowledges []model.Knowledge
		for records.Next(ctx) {
			node, _ := records.Record().Get("k")
			knowledges = append(knowledges, *r.nodeToKnowledge(node.(neo4j.Node)))
		}

		return knowledges, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]model.Knowledge), nil
}

func (r *knowledgeRepository) SearchByEmbedding(ctx context.Context, embedding []float64, limit int) ([]model.Knowledge, error) {
	session := r.session(ctx)
	defer session.Close(ctx)

	// 使用Neo4j向量索引（需要预先创建）
	cypher := `
		CALL db.index.vector.queryNodes('knowledge_embedding', $limit, $embedding)
		YIELD node, score
		RETURN node, score
		ORDER BY score DESC
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, cypher, map[string]interface{}{
			"embedding": embedding,
			"limit":     limit,
		})
		if err != nil {
			return nil, err
		}

		var knowledges []model.Knowledge
		for records.Next(ctx) {
			node, _ := records.Record().Get("node")
			knowledges = append(knowledges, *r.nodeToKnowledge(node.(neo4j.Node)))
		}

		return knowledges, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]model.Knowledge), nil
}

func (r *knowledgeRepository) GetRelated(ctx context.Context, id string, limit int) ([]model.Knowledge, error) {
	session := r.session(ctx)
	defer session.Close(ctx)

	cypher := `
		MATCH (k:Knowledge {id: $id})-[rel]-(related:Knowledge)
		RETURN related, type(rel) as relType
		LIMIT $limit
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, cypher, map[string]interface{}{
			"id":    id,
			"limit": limit,
		})
		if err != nil {
			return nil, err
		}

		var knowledges []model.Knowledge
		for records.Next(ctx) {
			node, _ := records.Record().Get("related")
			knowledges = append(knowledges, *r.nodeToKnowledge(node.(neo4j.Node)))
		}

		return knowledges, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]model.Knowledge), nil
}

func (r *knowledgeRepository) CreateRelation(ctx context.Context, relation *model.KnowledgeRelation) error {
	session := r.session(ctx)
	defer session.Close(ctx)

	cypher := fmt.Sprintf(`
		MATCH (source:Knowledge {id: $sourceId})
		MATCH (target:Knowledge {id: $targetId})
		CREATE (source)-[:%s {weight: $weight}]->(target)
	`, relation.RelationType)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		_, err := tx.Run(ctx, cypher, map[string]interface{}{
			"sourceId": relation.SourceID,
			"targetId": relation.TargetID,
			"weight":   relation.Weight,
		})
		return nil, err
	})

	return err
}

func (r *knowledgeRepository) GetGraph(ctx context.Context, subject, grade, userId string, limit int) (*model.KnowledgeGraph, error) {
	session := r.session(ctx)
	defer session.Close(ctx)

	// 只查询用户自己创建的知识点（通过文档上传生成）
	params := map[string]interface{}{
		"userId":  userId,
		"subject": subject,
		"grade":   grade,
		"limit":   int64(limit),
	}

	cypher := `
		MATCH (k:KnowledgePoint)
		WHERE k.userId = $userId
		  AND ($subject = '' OR k.subject = $subject OR k.subject IS NULL)
		  AND ($grade = '' OR k.grade CONTAINS $grade OR k.grade IS NULL)
		WITH k LIMIT $limit
		OPTIONAL MATCH (k)-[rel:DEPENDS_ON|RELATES_TO|SIMILAR_TO|PART_OF]-(related:KnowledgePoint)
		WHERE related.userId = $userId
		RETURN k, collect(DISTINCT {
			source: k.id,
			target: related.id,
			type: type(rel),
			weight: COALESCE(rel.strength, rel.similarity, 1.0)
		}) as relations
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		records, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}

		graph := &model.KnowledgeGraph{
			Nodes: []model.KnowledgeNode{},
			Edges: []model.KnowledgeEdge{},
		}

		edgeMap := make(map[string]bool)
		nodeMap := make(map[string]bool)

		for records.Next(ctx) {
			node, _ := records.Record().Get("k")
			neo4jNode := node.(neo4j.Node)
			props := neo4jNode.Props

			nodeID := ""
			if id, ok := props["id"].(string); ok {
				nodeID = id
			}
			if nodeID == "" {
				continue
			}

			// 避免重复添加节点
			if nodeMap[nodeID] {
				continue
			}
			nodeMap[nodeID] = true

			nodeName := ""
			if name, ok := props["name"].(string); ok {
				nodeName = name
			}

			nodeGrade := ""
			if g, ok := props["grade"].(string); ok {
				nodeGrade = g
			}

			nodeType := "KnowledgePoint"
			if t, ok := props["type"].(string); ok && t != "" {
				nodeType = t
			}

			nodeDifficulty := "medium"
			if d, ok := props["difficulty"].(string); ok && d != "" {
				nodeDifficulty = d
			}

			nodeImportance := 0.5
			if imp, ok := props["importance"].(float64); ok {
				nodeImportance = imp
			}

			nodeSubject := subject
			if s, ok := props["subject"].(string); ok && s != "" {
				nodeSubject = s
			}

			graph.Nodes = append(graph.Nodes, model.KnowledgeNode{
				ID:         nodeID,
				Label:      nodeName,
				Type:       nodeType,
				Subject:    nodeSubject,
				Grade:      nodeGrade,
				Difficulty: nodeDifficulty,
				Importance: nodeImportance,
			})

			relations, _ := records.Record().Get("relations")
			if rels, ok := relations.([]interface{}); ok {
				for _, rel := range rels {
					if relMap, ok := rel.(map[string]interface{}); ok {
						target, _ := relMap["target"].(string)
						if target == "" {
							continue
						}
						edgeKey := fmt.Sprintf("%s-%s", nodeID, target)
						reverseKey := fmt.Sprintf("%s-%s", target, nodeID)
						if edgeMap[edgeKey] || edgeMap[reverseKey] {
							continue
						}
						edgeMap[edgeKey] = true

						relType, _ := relMap["type"].(string)
						weight := 1.0
						if w, ok := relMap["weight"].(float64); ok {
							weight = w
						}

						graph.Edges = append(graph.Edges, model.KnowledgeEdge{
							Source: nodeID,
							Target: target,
							Type:   relType,
							Weight: weight,
						})
					}
				}
			}
		}

		// 过滤掉引用不存在节点的边（LIMIT 可能导致部分节点被裁剪）
		filteredEdges := make([]model.KnowledgeEdge, 0, len(graph.Edges))
		for _, edge := range graph.Edges {
			if nodeMap[edge.Source] && nodeMap[edge.Target] {
				filteredEdges = append(filteredEdges, edge)
			}
		}
		graph.Edges = filteredEdges

		return graph, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*model.KnowledgeGraph), nil
}

func (r *knowledgeRepository) nodeToKnowledge(node neo4j.Node) *model.Knowledge {
	props := node.Props

	k := &model.Knowledge{
		ID:   props["id"].(string),
		Name: props["name"].(string),
	}

	if v, ok := props["type"].(string); ok {
		k.Type = v
	}
	if v, ok := props["subject"].(string); ok {
		k.Subject = v
	}
	if v, ok := props["grade"].(string); ok {
		k.Grade = v
	}
	if v, ok := props["description"].(string); ok {
		k.Description = v
	}
	if v, ok := props["keywords"].([]interface{}); ok {
		for _, kw := range v {
			if s, ok := kw.(string); ok {
				k.Keywords = append(k.Keywords, s)
			}
		}
	}
	if embedding, ok := props["embedding"].([]interface{}); ok {
		for _, e := range embedding {
			if f, ok := e.(float64); ok {
				k.Embedding = append(k.Embedding, f)
			}
		}
	}

	return k
}
