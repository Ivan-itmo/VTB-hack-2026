export interface SourceItem {
  document_name: string;
  clause: string;
  edition_date: string;
  quote: string;
}

export interface CopilotCase {
  id: string;
  title: string;
  category: string;
  reg_code: string;
  incoming_task: string;
  suggested_route: string;
  confidence_score: number;
  escalation_required: boolean;
  resolution_draft: string;
  sources: SourceItem[];
}

export interface SandboxCase {
  id: string;
  title: string;
  incoming_task: string;
  sample_trainee_answer: string;
  score: number;
  is_correct: boolean;
  feedback: string;
  missed_rules: string[];
  correct_resolution: string;
}

export interface CasesData {
  copilot_cases: CopilotCase[];
  sandbox_cases: SandboxCase[];
}

export interface CopilotRequest {
  task_text: string;
  case_id?: string;
  category?: string;
}

export interface CopilotResponse {
  suggested_route: string;
  confidence_score: number;
  resolution_draft: string;
  escalation_required: boolean;
  sources: SourceItem[];
}

export interface SandboxRequest {
  task_text: string;
  trainee_solution: string;
  case_id?: string;
}

export interface SandboxResponse {
  is_correct: boolean;
  score: number;
  feedback: string;
  missed_rules: string[];
  correct_resolution: string;
}

export type AppMode = 'copilot' | 'sandbox';
