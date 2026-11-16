const { test, expect } = require('@playwright/test');

// Override this for another prepared calendar with $env:GUI_CALENDAR_USER_ID.
const ownerId = process.env.GUI_CALENDAR_USER_ID || 'bf18fa0e-e219-48f4-b586-1cef9b3b5a6c';
const calendarPath = `/book/${encodeURIComponent(ownerId)}`;

async function capture(page, testInfo, filename) {
  const screenshotPath = testInfo.outputPath(filename);
  await page.screenshot({ path: screenshotPath, fullPage: true });
  await testInfo.attach(filename, { path: screenshotPath, contentType: 'image/png' });
}

test.beforeEach(async ({ browser }, testInfo) => {
  testInfo.annotations.push({ type: 'browser', description: `Chromium ${browser.version()}` });
});

test('GUI-01: переход с главной страницы к форме входа', async ({ page }, testInfo) => {
  await test.step('Открыть главную страницу и проверить гостевой интерфейс', async () => {
    const response = await page.goto('/');
    expect(response.status()).toBe(200);
    await expect(page).toHaveTitle('ЛинкТайм');
    await expect(page.getByRole('heading', { name: 'Планируйте встречи проще' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Регистрация', exact: true })).toBeVisible();
    await capture(page, testInfo, '01-home.png');
  });

  await test.step('Нажать «Войти» и проверить форму входа', async () => {
    await page.getByRole('link', { name: 'Войти', exact: true }).click();
    await expect(page).toHaveURL(/\/api\/login$/);
    await expect(page.getByRole('heading', { name: 'Вход', exact: true })).toBeVisible();
    await expect(page.getByLabel('Email', { exact: true })).toBeEditable();
    await expect(page.getByLabel('Пароль', { exact: true })).toBeEditable();
    await expect(page.getByRole('button', { name: 'Войти', exact: true })).toBeEnabled();
    await capture(page, testInfo, '02-login.png');
  });
});

test('GUI-02: изменение даты публичного календаря', async ({ page }, testInfo) => {
  let selectedDate;

  await test.step('Открыть подготовленный календарь и проверить элементы управления', async () => {
    const response = await page.goto(calendarPath);
    expect(response.status(), 'Календарь должен существовать и иметь активное расписание').toBe(200);
    await expect(page.getByRole('heading', { name: /^Встреча с / })).toBeVisible();
    const dateInput = page.getByLabel('Дата встречи', { exact: true });
    await expect(dateInput).toBeEditable();
    await expect(page.getByRole('button', { name: 'Показать время', exact: true })).toBeEnabled();

    // Use the calendar's own date bounds, not a hard-coded date or the PC timezone.
    const min = await dateInput.getAttribute('min');
    const max = await dateInput.getAttribute('max');
    expect(min).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    const nextDay = new Date(`${min}T00:00:00Z`);
    nextDay.setUTCDate(nextDay.getUTCDate() + 1);
    selectedDate = nextDay.toISOString().slice(0, 10);
    expect(selectedDate <= max, 'Тестовая дата должна входить в разрешённый диапазон').toBe(true);
    expect(await dateInput.inputValue()).not.toBe(selectedDate);
  });

  await test.step('Выбрать другую дату и нажать «Показать время»', async () => {
    await page.getByLabel('Дата встречи', { exact: true }).fill(selectedDate);
    const navigation = page.waitForResponse(response => {
      const url = new URL(response.url());
      return response.request().isNavigationRequest()
        && response.request().method() === 'GET'
        && url.pathname === calendarPath
        && url.searchParams.get('date') === selectedDate;
    });
    await page.getByRole('button', { name: 'Показать время', exact: true }).click();
    expect((await navigation).status()).toBe(200);
  });

  await test.step('Проверить дату после перезагрузки страницы и результат запроса', async () => {
    await expect(page).toHaveURL(url => url.pathname === calendarPath && url.searchParams.get('date') === selectedDate);
    await expect(page.getByLabel('Дата встречи', { exact: true })).toHaveValue(selectedDate);
    await expect(page.getByRole('heading', { name: /^Встреча с / })).toBeVisible();
    await expect(page.getByRole('alert')).toHaveCount(0);
    // Both real schedule outcomes are valid: a list of slots or an empty-state message.
    await expect(page.locator('.booking-slots, .booking-empty')).toHaveCount(1);
    await expect(page.locator('.booking-slots, .booking-empty')).toBeVisible();
    await capture(page, testInfo, '03-calendar-date.png');
  });
});

test('GUI-03: некорректный email блокирует отправку формы регистрации', async ({ page }, testInfo) => {
  await test.step('Перейти к регистрации через ссылку на главной странице', async () => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Регистрация', exact: true }).click();
    await expect(page).toHaveURL(/\/api\/register$/);
    await expect(page.getByRole('heading', { name: 'Регистрация', exact: true })).toBeVisible();
  });

  await test.step('Заполнить форму с некорректным email', async () => {
    await page.getByLabel('Имя', { exact: true }).fill('GUI test');
    await page.getByLabel('Email', { exact: true }).fill('invalid-email');
    await page.getByLabel('Пароль', { exact: true }).fill('GuiTest-Password-123');
    await page.getByLabel('Повторите пароль', { exact: true }).fill('GuiTest-Password-123');
  });

  await test.step('Проверить встроенную валидацию и отсутствие отправки данных', async () => {
    let submissions = 0;
    page.on('request', request => {
      if (request.method() === 'POST' && new URL(request.url()).pathname === '/api/register') submissions++;
    });
    await page.getByRole('button', { name: 'Зарегистрироваться', exact: true }).click();
    const email = page.getByLabel('Email', { exact: true });
    await expect(email).toBeFocused();
    expect(await email.evaluate(input => input.validity.typeMismatch)).toBe(true);
    await expect(page).toHaveURL(/\/api\/register$/);
    expect(submissions, 'Браузер не должен отправлять форму с некорректным email').toBe(0);
    await capture(page, testInfo, '04-invalid-email.png');
  });
});
