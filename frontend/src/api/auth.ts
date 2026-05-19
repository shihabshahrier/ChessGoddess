// Auth API calls.
import { apiClient } from './client';

export async function getGoogleAuthURL(): Promise<{ url: string }> {
  const { data } = await apiClient.get('/auth/google/url');
  return data;
}
