// Shared TypeScript types matching backend model package.

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  created_at: string;
  updated_at: string;
}

export interface Game {
  id: string;
  user_id: string;
  pgn: string;
  fen: string;
  white_player: string;
  black_player: string;
  result: string;
  opening: string;
  time_control: string;
  event: string;
  date: string;
  created_at: string;
  updated_at: string;
}

export type MoveClassification = 'blunder' | 'mistake' | 'inaccuracy' | 'good' | 'excellent' | 'best';

export interface Move {
  id: string;
  session_id: string;
  move_number: number;
  fen: string;
  san: string;
  evaluation: number;
  best_move: string;
  classification: MoveClassification;
  depth: number;
  created_at: string;
}

export type AnalysisStatus = 'pending' | 'running' | 'completed' | 'failed';

export interface AnalysisSession {
  id: string;
  game_id: string;
  user_id: string;
  engine_config: string;
  depth: number;
  status: AnalysisStatus;
  started_at: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Snapshot {
  id: string;
  session_id: string;
  user_id: string;
  data: Record<string, unknown>;
  share_token: string;
  is_public: boolean;
  created_at: string;
}

export interface AIExplanation {
  id: string;
  session_id: string;
  move_id: string;
  fen: string;
  content: string;
  model: string;
  created_at: string;
}

export interface ApiError {
  error: string;
}
