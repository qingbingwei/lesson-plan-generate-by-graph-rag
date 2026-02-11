import { z } from 'zod';

// ==================== Skill Schema ====================

/**
 * Skill 定义 Schema
 * 每个 Skill 包含名称、描述和详细内容
 */
export const SkillSchema = z.object({
  name: z.string(),
  description: z.string(),
  content: z.string(),
});

export type Skill = z.infer<typeof SkillSchema>;

// ==================== 导出 Skills ====================

// 知识检索 Skill
export {
  knowledgeRetrievalSkill,
  retrieveKnowledge,
  getKnowledgeDetail,
} from './knowledgeRetrieval';

// 教学目标生成 Skill
export {
  objectiveGenerationSkill,
  generateObjectives,
  validateObjectives,
} from './objectiveGeneration';

// 教学内容生成 Skill
export {
  contentGenerationSkill,
  generateSections,
  generateKeyDifficultPoints,
  generateTeachingMethods,
  generateMaterials,
  generateHomework,
} from './contentGeneration';

// 评价设计 Skill
export {
  evaluationDesignSkill,
  generateEvaluation,
  generateRubric,
  generateFormativeAssessment,
  generateReflectionFramework,
} from './evaluationDesign';
