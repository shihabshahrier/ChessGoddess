// Auth API calls — Google OAuth flow.
import { apiClient } from './client'

export interface AuthUser {
  id: string
  email: string
  name: string
  avatar_url: string
}

/** Get the Google OAuth URL to redirect the browser to. */
export async function getGoogleAuthURL(): Promise<{ url: string }> {
  const { data } = await apiClient.get('/auth/google/url')
  return data
}

/** Fetch the currently signed-in user; rejects (401) when not signed in. */
export async function getMe(): Promise<AuthUser> {
  const { data } = await apiClient.get('/auth/me')
  return data.user
}

/** Clear the auth session cookie. */
export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout')
}
