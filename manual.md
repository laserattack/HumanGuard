
# Table of Contents

1.  [Запуск](#org82a0ab4)
2.  [Эндпоинты](#orgffee0ca)
    1.  [проверка что сервер работает](#org25e65dd)
    2.  [пользователи](#org8008658)
        1.  [создание пользователя](#orge835863)
        2.  [получение юзера по id](#org2ee0dc4)
        3.  [получение юзера по email](#org0db41f3)
        4.  [проверка занят ли email](#org411ee4a)
        5.  [смена пароля юзера](#orgbd48cec)
        6.  [обновление юзера по id](#orgaf3f9b2)
        7.  [обновление аватарки по id](#org7d36ebe)
        8.  [удаление юзера по id](#org2b50612)
        9.  [логин](#org842323f)
        10. [получение юзера по oauthId](#orgd50d96a)
    3.  [сайты](#org8593856)
        1.  [создание сайта](#org4e09a33)
        2.  [получение сайта](#org7c6b6f7)
        3.  [обновление сайта](#orgaf2f2f5)
        4.  [активация сайта](#org47f201c)
        5.  [блокировка сайта](#org72ae0bf)
        6.  [удаление сайта](#orge70d208)
        7.  [получение настроек сайта](#org9e91871)
        8.  [обновление настроек сайта](#orgf9c0c19)
    4.  [сессии](#org11e83f4)
        1.  [создать сессию](#orgd3ce253)
        2.  [получение сессии](#org3df81a2)
        3.  [получение сессии по cookie](#org4857cec)
        4.  [обновление сессии](#orgbeeb90d)
        5.  [обновить активность сессии (продлить жизнь)](#org89c9287)
        6.  [отметить что каптча показана](#orgc93fe39)
        7.  [Очистить просроченные сессии (админский)](#org7fe2572)
        8.  [обновление риск-скора](#orga1ec420)
        9.  [блокировка сессии](#org8e9c909)
        10. [разблокировка сессии](#org03bc72a)
        11. [деактивация сессии (удаление)](#org1718b0d)
        12. [активные сессии по сайту](#orge905773)
        13. [подозрительные сессии (риск более 60)](#org44b954c)
        14. [статистика по сайту](#orgf7e2eec)



<a id="org82a0ab4"></a>

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


<a id="orgffee0ca"></a>

# Эндпоинты


<a id="org25e65dd"></a>

## проверка что сервер работает

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/health
    {"status":"ok"}


<a id="org8008658"></a>

## пользователи


<a id="orge835863"></a>

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


<a id="org2ee0dc4"></a>

### получение юзера по id

GET api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:04:52.442724Z","last_login":null}


<a id="org0db41f3"></a>

### получение юзера по email

GET /api/users/email/{email}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/email/john@example.com
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Doe","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:07:02.521019Z","last_login":null}


<a id="org411ee4a"></a>

### проверка занят ли email

GET /api/users/exists?email=&#x2026;

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=john@example.com"
    {"exists":true}
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=xxx@example.com"
    {"exists":false}


<a id="orgbd48cec"></a>

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


<a id="orgaf3f9b2"></a>

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


<a id="org7d36ebe"></a>

### обновление аватарки по id

POST /api/users/{id}/avatar

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/avatar \
      -H "Content-Type: application/json" \
      -d '{
        "avatar_url": "https://example.com/avatars/user123.jpg"
      }'


<a id="org2b50612"></a>

### удаление юзера по id

DELETE api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    User not found


<a id="org842323f"></a>

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


<a id="orgd50d96a"></a>

### получение юзера по oauthId

GET /api/users/oauth/{provider}/{oauthId}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/oauth/google/123456789
    {"id":"d1e415b6-3855-4714-affc-7486231d0af6","email":"googleuser@example.com","name":"Google User","avatar_url":null,"role":"user","oauth_provider":"google","created_at":"2026-04-22T19:39:20.573219Z","updated_at":"2026-04-22T19:39:20.573219Z","last_login":null}


<a id="org8593856"></a>

## сайты


<a id="org4e09a33"></a>

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


<a id="org7c6b6f7"></a>

### получение сайта

GET /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:36:02.224274Z"}


<a id="orgaf2f2f5"></a>

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


<a id="org47f201c"></a>

### активация сайта

POST /api/sites/{id}/activate

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/activate


<a id="org72ae0bf"></a>

### блокировка сайта

POST /api/sites/{id}/suspend

    curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/suspend


<a id="orge70d208"></a>

### удаление сайта

DELETE /api/sites/{id}

    curl -X DELETE http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef


<a id="org9e91871"></a>

### получение настроек сайта

GET /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":false,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":""}}


<a id="orgf9c0c19"></a>

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


<a id="org11e83f4"></a>

## сессии


<a id="orgd3ce253"></a>

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


<a id="org3df81a2"></a>

### получение сессии

GET /api/sessions/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:31:40.920398Z"}


<a id="org4857cec"></a>

### получение сессии по cookie

GET /api/sessions/cookie/{cookie}

    curl http://localhost:8080/api/sessions/cookie/abc123


<a id="orgbeeb90d"></a>

### обновление сессии

PUT /api/sessions/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf \
      -H "Content-Type: application/json" \
      -d '{
        "ip": "10.0.0.1",
        "user_agent": "New Browser",
        "location": "SPb"
      }'


<a id="org89c9287"></a>

### обновить активность сессии (продлить жизнь)

POST /api/sessions/{id}/activity

    curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/activity


<a id="orgc93fe39"></a>

### отметить что каптча показана

POST /api/sessions/{id}/captcha

    curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/captcha


<a id="org7fe2572"></a>

### Очистить просроченные сессии (админский)

POST /api/sessions/cleanup

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/cleanup
    {"deleted":1}


<a id="orga1ec420"></a>

### обновление риск-скора

PATCH /api/sessions/{id}/risk

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PATCH http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/risk \
      -H "Content-Type: application/json" \
      -d '{"risk_score": 75}'


<a id="org8e9c909"></a>

### блокировка сессии

POST /api/sessions/{id}/block

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/block


<a id="org03bc72a"></a>

### разблокировка сессии

POST /api/sessions/{id}/unblock

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/unblock


<a id="org1718b0d"></a>

### деактивация сессии (удаление)

DELETE /api/sessions/{id}

    curl -X DELETE http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf


<a id="orge905773"></a>

### активные сессии по сайту

GET /api/sites/{id}/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions?limit=10"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org44b954c"></a>

### подозрительные сессии (риск более 60)

GET /api/sites/{id}/sessions/suspicious

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions/suspicious?min_risk=60"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="orgf7e2eec"></a>

### статистика по сайту

GET /api/sites/{id}/stats

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/stats
    {"total":1,"active":1,"blocked":0,"avg_risk":75,"unique_ips":1}
