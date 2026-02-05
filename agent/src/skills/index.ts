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

// ==================== Skills 列表 ====================

import { knowledgeRetrievalSkill } from './knowledgeRetrieval';
import { objectiveGenerationSkill } from './objectiveGeneration';
import { contentGenerationSkill } from './contentGeneration';
import { evaluationDesignSkill } from './evaluationDesign';

/**
 * 所有可用的 Skills 列表
 * 用于渐进式披露给 Agent
 */
export const SKILLS: Skill[] = [
  knowledgeRetrievalSkill,
  objectiveGenerationSkill,
  contentGenerationSkill,
  evaluationDesignSkill,
];

/**
 * 构建 Skills 提示词
 * 用于在系统提示词中展示可用技能
 */
export function buildSkillsPrompt(): string {
  return SKILLS.map((skill) => `- **${skill.name}**: ${skill.description}`).join('\n');
}

/**
 * 根据名称加载 Skill 内容
 */
export function loadSkill(skillName: string): string {
  const skill = SKILLS.find((s) => s.name === skillName);
  if (skill) {
    return `Loaded skill: ${skillName}\n\n${skill.content}`;
  }

  const available = SKILLS.map((s) => s.name).join(', ');
  return `Skill '${skillName}' not found. Available skills: ${available}`;
}
