# backend


# Start 
 
```
docker build -t backend .
docker run -p 8080:8080 backend
```

# Documentation
[API](API.md)



# Phantom MISIS: платформа анализа отзывов

Полный стек для загрузки CSV с отзывами, асинхронной обработки (классификация тональности, кластеризация, сводки кластеров) и визуализации результата. API на Go расставляет задачи в Celery/Redis, ML-воркер на Python готовит вывод, frontend на React/TanStack показывает аналитику.

## Ссылки
- Прод (Traefik): https://fcdc5ae656279f7ee87b25ea.duckdns.org
- Демо-видео: [Google Drive](https://drive.google.com/...) — замените на актуальную ссылку
- Репозитории: [backend](https://github.com/Phantom-misis/backend) · [frontend](https://github.com/Phantom-misis/frontend) · [ml](https://github.com/Phantom-misis/ML) · [deploy](https://github.com/Phantom-misis/deploy)

## Команда
- @okwaq — fullstack
- @ERR_4O4 — fullstack
- @Sofiia1398 — designer
- @ver_vina — ML
- @Timofey382 — ML

## Структура
- backend/ — Go 1.25.4 (Gin) API, Celery-клиент, маршруты /analyses, /reviews, /clusters
- frontend/ — React 19 + Vite 7 (FSD), TanStack Router/Query, HeroUI/Tailwind 4
- ml/ — Python 3.9+ ML-пайплайн: Celery-воркер (inference_worker.py), FastAPI (api_server.py), обучающие/оценочные скрипты
- docker-compose.yml — Traefik + Redis 8.4 + контейнер backend (ghcr.io/phantom-misis/backend)

## Технологии и версии
- Backend: Go 1.25.4, gin 1.11, gocelery, Redis, Traefik v3.6
- Frontend: Node 18+/pnpm 8+, React 19, TypeScript 5.7, Vite 7, TanStack Router/Query, HeroUI, Tailwind CSS 4, Vitest 3
- ML: Python 3.9+, PyTorch 2.5.1+cu121, Transformers 4.57.3, BERTopic, UMAP, SentenceTransformers, Celery, FastAPI

## Быстрый запуск (Docker Compose)
> В docker-compose.yml замените --requirepass ... на реальный пароль и создайте файл .env для backend.
1. Создайте .env рядом с docker-compose.yml:
      REDIS_HOST=redis
   REDIS_PORT=6379
   REDIS_PASSWORD=<ваш_пароль>
   
2. Поднимите инфраструктуру: docker compose up -d redis traefik backend
   - Redis будет на localhost:13394 (пароль как выше), backend — :8080 (проксируется Traefik на хосте fcdc5ae656279f7ee87b25ea.duckdns.org).
3. ML-воркер и frontend пока запускаются вручную (см. ниже) — добавьте их в compose при необходимости.

## Локальный запуск из исходников

### 1) Redis
docker run -d --name phantom-redis -p 6379:6379 -e REDIS_PASSWORD=12345678 redis:8.4.0 redis-server --requirepass 12345678

### 2) ML-воркер (Celery)
cd ml
python -m venv .venv && source .venv/bin/activate  # Windows: .venv\Scripts\activate
pip install -r requirements.txt --extra-index-url https://download.pytorch.org/whl/cu121
export REDIS_URL=redis://:12345678@localhost:6379/0
export MODEL_DIR=checkpoints  # путь к весам классификатора
celery -A inference_worker worker --loglevel=INFO
Дополнительно можно поднять HTTP API: uvicorn api_server:app --host 0.0.0.0 --port 8000.

### 3) Backend (Go API)
cd backend
set REDIS_HOST=localhost
set REDIS_PORT=6379
set REDIS_PASSWORD=12345678
go mod init backend   # однократно, если файла go.mod нет
go mod tidy
go run .
# или Docker: docker build -t phantom-backend . && docker run -p 8080:8080 --env-file ../.env phantom-backend
Сервер слушает :8080, публикует маршруты /analyses, /reviews, /clusters.

### 4) Frontend (React/Vite)
cd frontend
pnpm install
cat > .env <<'EOF'
VITE_API_URL=http://localhost:8080
VITE_GITHUB_URL=https://github.com/Phantom-misis/frontend
VITE_APP_VERSION=1.0.0
EOF
pnpm dev --host --port 3000
# сборка: pnpm build   # предпросмотр: pnpm serve
# тесты: pnpm test     # линт: pnpm lint
UI доступен на http://localhost:3000.

## API и формат данных
- Загрузка анализа: POST /analyses с формой file=<csv> (колонки text, src, опционально ID). Обрабатывается Celery-воркером, сохраняются метки настроения/уверенность, координаты UMAP, кластеры.
- Списки: GET /analyses, GET /analyses/:id, GET /analyses/:id/reviews, GET /analyses/:id/clusters.
- Редактирование метки: PATCH /reviews/:id ({"sentiment":"positive|neutral|negative"}).
- CSV ограничивается первыми 500 строками в воркере.
