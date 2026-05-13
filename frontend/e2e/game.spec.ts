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

async function startGame(browser: Browser): Promise<{
  code: string;
  ctx1: Awaited<ReturnType<Browser['newContext']>>;
  ctx2: Awaited<ReturnType<Browser['newContext']>>;
  page1: Page;
  page2: Page;
}> {
  const ctx1 = await browser.newContext();
  const page1 = await ctx1.newPage();
  const code = await createRoom(page1);

  const ctx2 = await browser.newContext();
  const page2 = await ctx2.newPage();
  await page2.goto(`/room/${code}`);
  await page2.getByPlaceholder('Ваше имя (до 40 символов)').fill('Игрок 2');
  await page2.getByRole('button', { name: 'Войти в комнату' }).click();

  await page1.waitForURL(new RegExp(`/game/${code}$`));
  await page2.waitForURL(new RegExp(`/game/${code}$`));
  await expect(page1.getByText('Загрузка доски...')).toBeHidden({ timeout: 10_000 });
  await expect(page2.getByText('Загрузка доски...')).toBeHidden({ timeout: 10_000 });

  return { code, ctx1, ctx2, page1, page2 };
}

test.describe('Full game lobby flow', () => {
  test('Two players can create and join a room', async ({ browser }) => {
    const { code, ctx1, ctx2 } = await startGame(browser);
    expect(code).toHaveLength(8);

    await ctx1.close();
    await ctx2.close();
  });

  test('Second player can join from copied room link', async ({ browser }) => {
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    const code = await createRoom(page1);

    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await page2.goto(`/room/${code}`);

    await page2.getByPlaceholder('Ваше имя (до 40 символов)').fill('Игрок 2');
    await page2.getByRole('button', { name: 'Войти в комнату' }).click();

    await page2.waitForURL(new RegExp(`/game/${code}$`));
    await page1.waitForURL(new RegExp(`/game/${code}$`));

    await ctx1.close();
    await ctx2.close();
  });

  test('Current player can see legal targets and move a checker', async ({ browser }) => {
    const { ctx1, ctx2, page1, page2 } = await startGame(browser);

    const currentPage = await page1.locator('.ring-2').filter({ hasText: '(вы)' }).count()
      ? page1
      : page2;
    const board = currentPage.getByTestId('game-board');
    const myColor = await board.getAttribute('data-my-color');
    const sourcePoint = myColor === 'white' ? '24' : '1';

    await currentPage.getByTestId(`point-${sourcePoint}`).click();
    const target = currentPage.getByTestId('valid-target').first();
    await expect(target).toBeVisible();
    await target.click();

    await expect(currentPage.getByTestId(`point-${sourcePoint}`)).toHaveAttribute(
      'data-checkers',
      '14',
    );
    await expect(currentPage.getByTestId('valid-target')).toHaveCount(0);

    await ctx1.close();
    await ctx2.close();
  });

  test('Waiting player cannot select checkers or see legal targets', async ({ browser }) => {
    const { ctx1, ctx2, page1, page2 } = await startGame(browser);

    const currentPage = await page1.locator('.ring-2').filter({ hasText: '(вы)' }).count()
      ? page1
      : page2;
    const waitingPage = currentPage === page1 ? page2 : page1;
    const board = waitingPage.getByTestId('game-board');
    const myColor = await board.getAttribute('data-my-color');
    const sourcePoint = myColor === 'white' ? '24' : '1';

    await waitingPage.getByTestId(`point-${sourcePoint}`).locator('circle').first().click();

    await expect(waitingPage.getByTestId('valid-target')).toHaveCount(0);

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
      test.skip(true, 'Backend not available');
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
