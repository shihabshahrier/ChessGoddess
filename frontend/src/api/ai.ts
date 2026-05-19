// AI explanation API calls.
import type { AIExplanation } from '../types';
import { apiClient } from './client';

export async function explainMove(moveId: string, sessionId: string): Promise<{ explanation: string }> {
  const { data } = await apiClient.post('/ai/explain', { move_id: moveId, session_id: sessionId });
  return data;
}

export async function explainBlunder(moveId: string, sessionId: string): Promise<{ explanation: string }> {
  const { data } = await apiClient.post('/ai/explain-blunder', { move_id: moveId, session_id: sessionId });
  return data;
}

export async function getExplanation(moveId: string): Promise<{ explanation: AIExplanation }> {
  const { data } = await apiClient.get(`/ai/explanation/${moveId}`);
  return data;
}
