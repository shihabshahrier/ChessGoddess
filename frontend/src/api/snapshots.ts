// Snapshot API calls.
import type { Snapshot } from '../types';
import { apiClient } from './client';

export async function createSnapshot(sessionId: string): Promise<{ session_id: string }> {
  const { data } = await apiClient.post(`/snapshots?session_id=${sessionId}`);
  return data;
}

export async function listSnapshots(): Promise<{ snapshots: Snapshot[] }> {
  const { data } = await apiClient.get('/snapshots');
  return data;
}

export async function getSnapshotByToken(token: string): Promise<{ snapshot: Snapshot }> {
  const { data } = await apiClient.get(`/snapshots/${token}`);
  return data;
}
