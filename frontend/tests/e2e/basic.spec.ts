import { test, expect } from '@playwright/test'

test.describe('ChessLens Home Page', () => {
  test('should load the home page', async ({ page }) => {
    await page.goto('/')
    
    await expect(page).toHaveTitle(/ChessLens/)
    
    const heading = page.getByRole('heading', { name: /See Chess/ })
    await expect(heading).toBeVisible()
  })

  test('should display feature cards', async ({ page }) => {
    await page.goto('/')
    
    await expect(page.getByText('Deep Analysis')).toBeVisible()
    await expect(page.getByText('Cinematic Review')).toBeVisible()
    await expect(page.getByText('AI Explanations')).toBeVisible()
  })

  test('should navigate to upload page', async ({ page }) => {
    await page.goto('/')
    
    await page.getByRole('link', { name: 'Upload Game' }).click()
    
    await expect(page).toHaveURL('/upload')
    await expect(page.getByRole('heading', { name: 'Upload a Game' })).toBeVisible()
  })
})

test.describe('Upload Page', () => {
  test('should show upload interface', async ({ page }) => {
    await page.goto('/upload')
    
    await expect(page.getByRole('heading', { name: 'Upload a Game' })).toBeVisible()
    await expect(page.getByRole('textbox', { name: /paste PGN/i })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Analyze Game' })).toBeVisible()
  })

  test('should disable analyze button without PGN', async ({ page }) => {
    await page.goto('/upload')
    
    const analyzeButton = page.getByRole('button', { name: 'Analyze Game' })
    await expect(analyzeButton).toBeDisabled()
  })
})

test.describe('Share Page', () => {
  test('should show shared analysis', async ({ page }) => {
    await page.goto('/s/test-snapshot-id')
    
    await expect(page.getByRole('heading', { name: 'Shared Analysis' })).toBeVisible()
    await expect(page.getByText(/immutable snapshot/i)).toBeVisible()
  })
})
