// Game API calls.
import type { Game, AnalysisSession, Move } from '../types';
import { apiClient } from './client';

export async function uploadGame(pgn: string): Promise<{ game_id: string }> {
  const { data } = await apiClient.post('/games/upload', { pgn });
  return data;
}

export async function listGames(): Promise<{ games: Game[] }> {
  const { data } = await apiClient.get('/games');
  return data;
}

export async function getGame(id: string): Promise<{ game: Game }> {
  const { data } = await apiClient.get(`/games/${id}`);
  return data;
}

export async function deleteGame(id: string): Promise<void> {
  await apiClient.delete(`/games/${id}`);
}

export async function createAnalysis(gameId: string, depth = 20): Promise<{ session_id: string; status: string }> {
  const { data } = await apiClient.post('/analysis', { game_id: gameId, depth });
  return data;
}

export async function getAnalysis(sessionId: string): Promise<{ session: AnalysisSession }> {
  const { data } = await apiClient.get(`/analysis/${sessionId}`);
  return data;
}

export async function getAnalysisMoves(sessionId: string): Promise<{ moves: Move[] }> {
  const { data } = await apiClient.get(`/analysis/${sessionId}/moves`);
  return data;
}
