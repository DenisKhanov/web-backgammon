import { test, expect, Browser, Page } from '@playwright/test';

async function createRoom(page: Page): Promise<string> {
  await page.goto('/');
  await page.getByPlaceholder('Ваше имя (до 40 символов)').first().fill('Игрок 1');
  await page.getByRole('button', { name: 'Создать комнату' }).click();
  // Wait for /room/[code] redirect
  await page.waitForURL(/\/room\//);
  const code = page.url().split('/').pop()!.toUpperCase();
  return code;
}

test.describe('Full game lobby flow', () => {
  test('Two players can create and join a room', async ({ browser }) => {
    // Player 1 creates room
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    const code = await createRoom(page1);
    expect(code).toHaveLength(8);

    // Player 2 joins room
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await page2.goto('/');
    await page2.getByPlaceholder('Код комнаты (8 символов)').fill(code);
    await page2.getByPlaceholder('Ваше имя (до 40 символов)').last().fill('Игрок 2');
    await page2.getByRole('button', { name: 'Войти' }).click();

    // Both pages redirect to /room/[code]
    await page2.waitForURL(/\/room\//);
    expect(page2.url()).toContain(code);

    await ctx1.close();
    await ctx2.close();
  });

  test('Landing page renders correctly on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/');
    await expect(page.getByText('Длинные нарды')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Создать комнату' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Войти' })).toBeVisible();
  });

  test('Room page shows code and copy link button', async ({ page }) => {
    // This test requires a backend. Skip if not available.
    const resp = await page.request.get('http://localhost:8080/api/health').catch(() => null);
    if (!resp || !resp.ok()) {
      test.skip();
      return;
    }

    await page.goto('/');
    await page.getByPlaceholder('Ваше имя (до 40 символов)').first().fill('TestPlayer');
    await page.getByRole('button', { name: 'Создать комнату' }).click();
    await page.waitForURL(/\/room\//);

    await expect(page.getByText(/[A-Z2-7]{8}/)).toBeVisible();
    await expect(page.getByText('Скопировать ссылку')).toBeVisible();
  });
});
