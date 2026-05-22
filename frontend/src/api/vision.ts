// Vision API calls — chess board image to FEN.
import { apiClient } from './client'

/** Upload a board image (PNG/JPEG/WebP) and get back the recognized FEN. */
export async function imageToFEN(file: File): Promise<{ fen: string }> {
  const form = new FormData()
  form.append('image', file)
  // axios strips the JSON default and sets multipart + boundary for FormData.
  const { data } = await apiClient.post('/vision/image-to-fen', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data
}

/** Recognize a board from a remote image URL. */
export async function imageURLToFEN(imageURL: string): Promise<{ fen: string }> {
  const { data } = await apiClient.post('/vision/image-to-fen-url', {
    image_url: imageURL,
  })
  return data
}
