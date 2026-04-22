
# Table of Contents

1.  [Запуск](#orgd449bee)
2.  [Эндпоинты](#orgb106f76)
    1.  [проверка что сервер работает](#org5125f7a)
    2.  [пользователи](#orgb5034f9)
        1.  [создание пользователя](#org90149fe)
        2.  [получение юзера по id](#orgd4b521c)
        3.  [получение юзера по email](#org5586851)
        4.  [проверка занят ли email](#org26b35e9)
        5.  [смена пароля юзера](#orgb176abb)
        6.  [обновление юзера по id](#org6455b80)
        7.  [обновление аватарки по id](#org129987a)
        8.  [удаление юзера по id](#org3fa9f95)
        9.  [логин](#orgbf3a85a)
    3.  [сайты](#orge0df258)
        1.  [создание сайта](#org300a7f1)
        2.  [получение сайта](#org746b9af)
        3.  [обновление сайта](#org07a391d)
        4.  [активация сайта](#org8e9233b)
        5.  [блокировка сайта](#org8370305)
        6.  [удаление сайта](#org8357420)
        7.  [получение настроек сайта](#org40daba5)
        8.  [обновление настроек сайта](#org5d2b443)
    4.  [сессии](#org9cae4b6)
        1.  [создать сессию](#orgf072c59)
        2.  [получение сессии](#org3b72598)
        3.  [обновление риск-скора](#orgfb4c9ea)
        4.  [блокировка сессии](#org1287939)
        5.  [разблокировка сессии](#org554bf31)
        6.  [деактивация сессии (удаление)](#org558e79e)
        7.  [активные сессии по сайту](#org8b7cb1f)
        8.  [подозрительные сессии (риск более 60)](#orgec03ebe)
        9.  [статистика по сайту](#org2b087bd)



<a id="orgd449bee"></a>

# Запуск

база

    docker run -d \
      --name humanguard-db \
      -e POSTGRES_DB=humanguard \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=123 \
      -p 5432:5432 \
      postgres:15

применяю миграции и запускаю прилку (из директории backend)

    docker cp migrations/001_init_up.sql humanguard-db:/tmp/init.sql
    docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/001_init_up.sql
    docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/002_add_oauth_totp_up.sql
    go run cmd/server/main.go


<a id="orgb106f76"></a>

# Эндпоинты


<a id="org5125f7a"></a>

## проверка что сервер работает

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/health
    {"status":"ok"}


<a id="orgb5034f9"></a>

## пользователи


<a id="org90149fe"></a>

### создание пользователя

POST api/users

все поля являются обязательными

    var req struct {
        Email        string `json:"email"`
        Name         string `json:"name"`
        PasswordHash string `json:"password_hash"`
        Role         string `json:"role"`
    }

пример

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users   -H "Content-Type: application/json"   -d '{
        "email": "test@example.com",
        "name": "Test User",
        "password_hash": "hash123",
        "role": "user"
      }'
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724101+03:00","updated_at":"2026-04-22T19:04:52.442724101+03:00","last_login":null}


<a id="orgd4b521c"></a>

### получение юзера по id

GET api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:04:52.442724Z","last_login":null}


<a id="org5586851"></a>

### получение юзера по email

GET /api/users/email/{email}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/email/john@example.com
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Doe","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:07:02.521019Z","last_login":null}


<a id="org26b35e9"></a>

### проверка занят ли email

GET /api/users/exists?email=&#x2026;

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=john@example.com"
    {"exists":true}
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=xxx@example.com"
    {"exists":false}


<a id="orgb176abb"></a>

### смена пароля юзера

POST /api/users/{id}/password

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/password \
      -H "Content-Type: application/json" \
      -d '{
        "old_password": "secret123",
        "new_password": "newpassword456"
      }'
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/password \
      -H "Content-Type: application/json" \
      -d '{
        "old_password": "secret123",
        "new_password": "newpassword456"
      }'
    Invalid old password


<a id="org6455b80"></a>

### обновление юзера по id

PUT api/users/{id}

вот такие поля можно обновлять. можно сразу несколько, можно и по одному

    var req struct {
        Name string `json:"name"`
        Role string `json:"role"`
    }

пример

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2 \
      -H "Content-Type: application/json" \
      -d '{
        "name": "New Name",
        "role": "admin"
      }'
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"New Name","avatar_url":null,"role":"admin","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:28:34.916821321+03:00","last_login":null}


<a id="org129987a"></a>

### обновление аватарки по id

POST /api/users/{id}/avatar

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/avatar \
      -H "Content-Type: application/json" \
      -d '{
        "avatar_url": "https://example.com/avatars/user123.jpg"
      }'


<a id="org3fa9f95"></a>

### удаление юзера по id

DELETE api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    User not found


<a id="orgbf3a85a"></a>

### логин

POST /api/login

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "john@example.com",
        "password": "secret123"
      }'
    Invalid credentials
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "john@example.com",
        "password": "newpassword456"
      }'
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Updated","avatar_url":null,"role":"admin","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:09:48.553076Z","last_login":null}


<a id="orge0df258"></a>

## сайты


<a id="org300a7f1"></a>

### создание сайта

POST /api/sites

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "aa43b9e7-5142-49bd-8e6c-c4e6745e91f7",
        "name": "My Blog",
        "domain": "blog.example.com",
        "origin_server": "https://origin-blog.example.com"
      }'
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":null,"created_at":"2026-04-22T21:36:02.224274333+03:00","updated_at":"2026-04-22T21:36:02.224274333+03:00"}


<a id="org746b9af"></a>

### получение сайта

GET /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:36:02.224274Z"}


<a id="org07a391d"></a>

### обновление сайта

PUT /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Updated Blog",
        "status": "active"
      }'
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"Updated Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"active","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:38:57.180079014+03:00"}


<a id="org8e9233b"></a>

### активация сайта

POST /api/sites/{id}/activate

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/activate


<a id="org8370305"></a>

### блокировка сайта

POST /api/sites/{id}/suspend

    curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/suspend


<a id="org8357420"></a>

### удаление сайта

DELETE /api/sites/{id}

    curl -X DELETE http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef


<a id="org40daba5"></a>

### получение настроек сайта

GET /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":false,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":""}}


<a id="org5d2b443"></a>

### обновление настроек сайта

PUT /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings \
      -H "Content-Type: application/json" \
      -d '{
        "collector": {
          "enabled": true,
          "mouse_tracking": true,
          "click_tracking": true,
          "scroll_tracking": true,
          "keystroke_tracking": false,
          "fingerprint_enabled": true
        },
        "analyzer": {
          "enabled": true,
          "rate_limiting": true,
          "pattern_analysis": true,
          "headless_detection": true,
          "thresholds": {
            "low": 30,
            "medium": 60,
            "high": 80
          }
        },
        "reaction": {
          "enabled": true,
          "low_risk_action": "allow",
          "medium_risk_action": "captcha",
          "high_risk_action": "block",
          "block_duration": 60,
          "captcha_provider": "hcaptcha"
        }
      }'
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":true,"keystroke_tracking":false,"fingerprint_enabled":true},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":true,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"allow","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":"hcaptcha"}}


<a id="org9cae4b6"></a>

## сессии


<a id="orgf072c59"></a>

### создать сессию

POST /api/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions \
      -H "Content-Type: application/json" \
      -d '{
        "site_id": "138fba32-f0a9-43ce-8a52-188cd721c2ef",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0",
        "device": "desktop",
        "location": "Moscow"
      }'
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920397824+03:00","last_activity":"2026-04-22T22:01:40.920397824+03:00","expires_at":"2026-04-22T22:31:40.920397824+03:00"}


<a id="org3b72598"></a>

### получение сессии

GET /api/sessions/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:31:40.920398Z"}


<a id="orgfb4c9ea"></a>

### обновление риск-скора

PATCH /api/sessions/{id}/risk

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PATCH http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/risk \
      -H "Content-Type: application/json" \
      -d '{"risk_score": 75}'


<a id="org1287939"></a>

### блокировка сессии

POST /api/sessions/{id}/block

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/block


<a id="org554bf31"></a>

### разблокировка сессии

POST /api/sessions/{id}/unblock

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/unblock


<a id="org558e79e"></a>

### деактивация сессии (удаление)

DELETE /api/sessions/{id}

    curl -X DELETE http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf


<a id="org8b7cb1f"></a>

### активные сессии по сайту

GET /api/sites/{id}/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions?limit=10"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="orgec03ebe"></a>

### подозрительные сессии (риск более 60)

GET /api/sites/{id}/sessions/suspicious

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions/suspicious?min_risk=60"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org2b087bd"></a>

### статистика по сайту

GET /api/sites/{id}/stats

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/stats
    {"total":1,"active":1,"blocked":0,"avg_risk":75,"unique_ips":1}
