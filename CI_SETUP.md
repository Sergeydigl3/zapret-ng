# Настройка CI/CD

Этот документ описывает настройку CI/CD для автоматической сборки и публикации пакетов zapret-daemon.

## Обзор

Проект использует:
- **GitHub Actions** - для автоматизации CI/CD
- **GoReleaser** - для сборки кросс-платформенных пакетов
- Поддержка форматов: **DEB**, **RPM**, **APK**, **Arch Linux**, **tar.gz**
- Поддержка систем инициализации: **systemd**, **OpenRC**, **SysVinit**

## Структура файлов

```
zapret-discord-youtube-ng/
├── .github/workflows/
│   └── release.yml              # GitHub Actions workflow
├── .goreleaser.yml              # Конфигурация GoReleaser
├── init/
│   ├── systemd/
│   │   └── zapret-daemon.service
│   ├── openrc/
│   │   └── zapret-daemon
│   └── sysvinit/
│       └── zapret-daemon
├── scripts/
│   ├── postinstall.sh           # Скрипт после установки
│   └── preremove.sh             # Скрипт перед удалением
├── out/
│   ├── bin/                     # Локальная сборка бинарников
│   └── dist/                    # GoReleaser артефакты
└── CI_SETUP.md                  # Этот файл
```

## Первоначальная настройка

### 1. GitHub Repository

Убедитесь, что репозиторий создан на GitHub:
```bash
git remote -v
```

### 2. Права доступа

GitHub Actions уже имеет необходимые права через `GITHUB_TOKEN`, который автоматически предоставляется.

Проверьте настройки репозитория:
1. Перейдите в **Settings** → **Actions** → **General**
2. В разделе **Workflow permissions** выберите:
   - ✅ Read and write permissions
   - ✅ Allow GitHub Actions to create and approve pull requests

### 3. Проверка конфигурации

Перед первым релизом проверьте конфигурацию локально:

```bash
# Установить GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Проверить конфигурацию
goreleaser check

# Собрать снапшот (без публикации)
make proto  # Сгенерировать protobuf код
goreleaser release --snapshot --clean

# Проверить результат
ls -la out/dist/
```

## Создание релиза

### Процесс релиза

1. **Убедитесь, что код готов к релизу:**
   ```bash
   # Проверить статус
   git status

   # Все тесты проходят
   go test ./...

   # Код собирается
   make build
   ```

2. **Создайте и запушьте тег:**
   ```bash
   # Создать тег с аннотацией
   git tag -a v1.0.0 -m "Release v1.0.0

   Features:
   - Поддержка RPC через Twirp
   - Конфигурация через YAML
   - Системные сервисы для всех init систем
   "

   # Проверить тег
   git tag -l -n9 v1.0.0

   # Запушить тег
   git push origin v1.0.0
   ```

3. **GitHub Actions автоматически:**
   - Соберет код
   - Создаст пакеты для всех платформ
   - Опубликует релиз
   - Загрузит артефакты

4. **Проверьте релиз:**
   - Перейдите в **Releases** на GitHub
   - Убедитесь, что все пакеты созданы
   - Проверьте checksums
   - Протестируйте установку на целевых платформах

## Workflow детали

### Триггеры

Workflow запускается при:
- Пуше тега `v*` (например, `v1.0.0`, `v2.1.3`)
- Ручном запуске через GitHub UI (для тестирования)

### Этапы сборки

1. **Checkout** - клонирование репозитория
2. **Setup Go** - установка Go 1.21
3. **Install protoc** - установка Protocol Buffers compiler
4. **Install Go tools** - установка protoc-gen-go и protoc-gen-twirp
5. **Generate protobuf** - генерация RPC кода
6. **GoReleaser** - сборка и публикация

### Архитектуры

Собираются пакеты для:
- **amd64** (x86_64) - для большинства десктопов и серверов
- **arm64** (aarch64) - для ARM64 серверов, Raspberry Pi 3+, Apple Silicon
- **arm** v6/v7 (armhf) - для старых ARM устройств, Raspberry Pi 1/2
- **386** (i386) - для старых 32-битных систем

## Пакеты

### DEB (Debian/Ubuntu)

- Формат: `zapret-daemon_<version>_<arch>.deb`
- Init система: systemd или SysVinit (автоопределение)
- Зависимости: iptables
- Скрипты: postinstall, preremove

### RPM (RHEL/CentOS/Fedora)

- Формат: `zapret-daemon-<version>-1.<arch>.rpm`
- Init система: systemd
- Зависимости: iptables
- Скрипты: post, preun

### APK (Alpine Linux)

- Формат: `zapret-daemon-<version>-r0.<arch>.apk`
- Init система: OpenRC
- Зависимости: iptables, openrc
- Скрипты: post-install, pre-deinstall

### Arch Linux

- Формат: `zapret-daemon-<version>-1-<arch>.pkg.tar.zst`
- Init система: systemd
- Зависимости: iptables, systemd

### TAR.GZ (универсальный)

- Формат: `zapret-daemon_<version>_<os>_<arch>.tar.gz`
- Содержит: бинарники, конфиги, init скрипты
- Требует ручной установки

## Скрипты установки

### postinstall.sh

Выполняется после установки пакета:
1. Создает runtime директорию `/var/run/zapret`
2. Определяет систему инициализации (systemd/OpenRC/SysVinit)
3. Регистрирует и включает сервис
4. Выводит инструкции по запуску

### preremove.sh

Выполняется перед удалением пакета:
1. Останавливает сервис
2. Отключает автозапуск
3. Очищает PID файлы
4. Удаляет символические ссылки

## Тестирование

### Локальное тестирование сборки

```bash
# Полная сборка пакетов
goreleaser release --snapshot --clean

# Проверить содержимое DEB
dpkg-deb -c out/dist/zapret-daemon_*_amd64.deb

# Проверить содержимое RPM
rpm -qlp out/dist/zapret-daemon-*-1.x86_64.rpm

# Проверить содержимое TAR.GZ
tar -tzf out/dist/zapret-daemon_*_linux_amd64.tar.gz
```

### Тестирование установки

#### Docker тестирование

```bash
# Debian/Ubuntu
docker run -it --rm -v $(pwd)/out/dist:/dist debian:bookworm bash
apt update && apt install -y /dist/zapret-daemon_*_amd64.deb
systemctl status zapret-daemon || service zapret-daemon status

# Alpine Linux
docker run -it --rm -v $(pwd)/out/dist:/dist alpine:latest sh
apk add --allow-untrusted /dist/zapret-daemon-*-r0.x86_64.apk
rc-service zapret-daemon status

# RHEL/CentOS
docker run -it --rm -v $(pwd)/out/dist:/dist rockylinux:9 bash
dnf install -y /dist/zapret-daemon-*-1.x86_64.rpm
systemctl status zapret-daemon
```

## Отладка

### Просмотр логов GitHub Actions

1. Перейдите в **Actions** → выберите workflow run
2. Кликните на job "release" или "build-test"
3. Раскройте нужный step для просмотра логов

### Частые проблемы

#### 1. Ошибка "protoc not found"

Убедитесь, что protoc установлен в workflow:
```yaml
- name: Install protoc
  run: |
    PROTOC_VERSION=28.3
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip
    ...
```

#### 2. Ошибка "permission denied" при публикации

Проверьте permissions в workflow:
```yaml
permissions:
  contents: write
  packages: write
```

#### 3. GoReleaser ошибки

Проверьте конфигурацию:
```bash
goreleaser check
```

#### 4. Init скрипты не работают

Убедитесь, что файлы имеют правильные права:
```bash
chmod +x init/openrc/zapret-daemon
chmod +x init/sysvinit/zapret-daemon
```

## Версионирование

Проект использует Semantic Versioning (semver):

- **MAJOR** (1.x.x) - несовместимые изменения API
- **MINOR** (x.1.x) - новая функциональность, обратная совместимость
- **PATCH** (x.x.1) - исправления багов

Примеры:
```bash
git tag -a v1.0.0 -m "Initial release"
git tag -a v1.1.0 -m "Add new features"
git tag -a v1.1.1 -m "Bug fixes"
git tag -a v2.0.0 -m "Breaking changes"
```

## Changelog

GoReleaser автоматически генерирует changelog из коммитов.

Используйте Conventional Commits для лучшей группировки:

```bash
feat: добавить поддержку конфигурации через TOML
fix: исправить утечку памяти в RPC сервере
perf: оптимизировать обработку сообщений
docs: обновить README
chore: обновить зависимости
```

## Безопасность

### Подпись пакетов (опционально)

Для подписи пакетов GPG ключом:

1. Создайте GPG ключ:
   ```bash
   gpg --full-generate-key
   ```

2. Экспортируйте приватный ключ:
   ```bash
   gpg --armor --export-secret-keys YOUR_KEY_ID > private.key
   ```

3. Добавьте секрет в GitHub:
   - Settings → Secrets → Actions
   - Создайте секрет `GPG_PRIVATE_KEY` с содержимым private.key

4. Обновите .goreleaser.yml:
   ```yaml
   signs:
     - artifacts: checksum
       args: ["--batch", "-u", "{{ .Env.GPG_FINGERPRINT }}", "--output", "${signature}", "--detach-sign", "${artifact}"]
   ```

## Следующие шаги

1. ✅ Создать первый релиз: `v1.0.0`
2. ✅ Протестировать установку на разных дистрибутивах
3. Настроить автоматические тесты
4. Добавить поддержку других архитектур (mips, ppc64le)
5. Настроить репозитории пакетов (APT, YUM, AUR)
6. Добавить Docker образы

## Полезные ссылки

- [GoReleaser Documentation](https://goreleaser.com/intro/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [NFPM (пакеты)](https://nfpm.goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
