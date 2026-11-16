const fs = require('node:fs');
const path = require('node:path');

class LabReporter {
  onBegin(config, suite) {
    this.started = new Date();
    this.total = suite.allTests().length;
    this.results = [];
    this.baseURL = config.projects[0].use.baseURL;
  }

  onTestEnd(test, result) {
    this.results.push({
      title: test.title,
      status: result.status,
      duration: result.duration,
      browser: test.annotations.find(annotation => annotation.type === 'browser')?.description || 'Chromium',
      errors: result.errors.map(error => error.message || String(error)),
    });
  }

  onEnd(result) {
    const output = path.resolve(__dirname, '../results/gui');
    fs.mkdirSync(output, { recursive: true });
    const passed = this.results.filter(test => test.status === 'passed').length;
    const lines = [
      'GUI-тестирование веб-сервиса «ЛинкТайм»',
      `Начало: ${this.started.toISOString()}`,
      `Сервис: ${this.baseURL}`,
      `Среда: Node.js ${process.version}; ${process.platform}/${process.arch}`,
      `Результат запуска: ${result.status}`,
      `Всего тестов: ${this.total}; успешно: ${passed}; остальные: ${this.total - passed}`,
      '',
    ];
    for (const test of this.results) {
      lines.push(`${test.title}: ${test.status}; ${test.duration} мс; ${test.browser}`);
      lines.push(...test.errors);
    }
    lines.push('', 'Проверки выполнены в настоящем браузере против работающего сервиса, без подмены HTTP-ответов.',
      'Каждый тест использует отдельный браузерный контекст без сохранённой авторизации.',
      'HTML-отчёт: playwright-report/index.html. Скриншоты и трассировки доступны во вложениях тестов.',
      'Продолжительность теста включает действия и проверки; это не показатель производительности сервера.');
    fs.writeFileSync(path.join(output, 'run-summary.txt'), `${lines.join('\n')}\n`, 'utf8');
  }
}

module.exports = LabReporter;
