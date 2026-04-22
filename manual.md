
# Table of Contents

1.  [Запуск](#org067f479)
2.  [Эндпоинты](#orga968871)
    1.  [проверка что сервер работает](#orga658a77)
    2.  [пользователи](#org65f217a)
        1.  [создание пользователя](#org8dc07bb)
        2.  [получение юзера по id](#orgb6a04db)
        3.  [получение юзера по email](#orge71de3c)
        4.  [проверка занят ли email](#orgef5e1bc)
        5.  [смена пароля юзера](#org2b84212)
        6.  [обновление юзера по id](#orgdc235ad)
        7.  [удаление юзера по id](#org0c2f8ff)
        8.  [логин](#org3ddb80d)
    3.  [сайты](#orgfab5ba6)
        1.  [создание сайта](#orgbd7337e)
        2.  [получение сайта](#org1a2510a)
        3.  [обновление сайта](#org6112e40)
        4.  [активация сайта](#orge22e976)
        5.  [блокировка сайта](#org1be7810)
        6.  [удаление сайта](#orgb7ba4c4)
    4.  [сессии](#orgac51f51)
        1.  [создать сессию](#org0819e03)
        2.  [получение сессии](#orgffaf3ce)
        3.  [обновление риск-скора](#org33314c5)
        4.  [блокировка сессии](#org1c2e963)
        5.  [разблокировка сессии](#org2e46124)
        6.  [деактивация сессии (удаление)](#orgf7b0859)
        7.  [активные сессии по сайту](#org6af6e4d)
        8.  [подозрительные сессии (риск более 60)](#orgc44966c)
        9.  [статистика по сайту](#orgee3941d)



<a id="org067f479"></a>

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


<a id="orga968871"></a>

# Эндпоинты


<a id="orga658a77"></a>

## проверка что сервер работает

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/health
    {"status":"ok"}


<a id="org65f217a"></a>

## пользователи


<a id="org8dc07bb"></a>

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


<a id="orgb6a04db"></a>

### получение юзера по id

GET api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:04:52.442724Z","last_login":null}


<a id="orge71de3c"></a>

### получение юзера по email

GET /api/users/email/{email}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/email/john@example.com
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Doe","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:07:02.521019Z","last_login":null}


<a id="orgef5e1bc"></a>

### проверка занят ли email

GET /api/users/exists?email=&#x2026;

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=john@example.com"
    {"exists":true}
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=xxx@example.com"
    {"exists":false}


<a id="org2b84212"></a>

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


<a id="orgdc235ad"></a>

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


<a id="org0c2f8ff"></a>

### удаление юзера по id

DELETE api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    User not found


<a id="org3ddb80d"></a>

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


<a id="orgfab5ba6"></a>

## сайты


<a id="orgbd7337e"></a>

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


<a id="org1a2510a"></a>

### получение сайта

GET /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:36:02.224274Z"}


<a id="org6112e40"></a>

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


<a id="orge22e976"></a>

### активация сайта

POST /api/sites/{id}/activate

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/activate


<a id="org1be7810"></a>

### блокировка сайта

POST /api/sites/{id}/suspend

    curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/suspend


<a id="orgb7ba4c4"></a>

### удаление сайта

DELETE /api/sites/{id}

    curl -X DELETE http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef


<a id="orgac51f51"></a>

## сессии


<a id="org0819e03"></a>

### создать сессию

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


<a id="orgffaf3ce"></a>

### получение сессии

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:31:40.920398Z"}


<a id="org33314c5"></a>

### обновление риск-скора

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PATCH http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/risk \
      -H "Content-Type: application/json" \
      -d '{"risk_score": 75}'


<a id="org1c2e963"></a>

### блокировка сессии

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/block


<a id="org2e46124"></a>

### разблокировка сессии

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/unblock


<a id="orgf7b0859"></a>

### деактивация сессии (удаление)

    curl -X DELETE http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf


<a id="org6af6e4d"></a>

### активные сессии по сайту

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions?limit=10"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="orgc44966c"></a>

### подозрительные сессии (риск более 60)

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions/suspicious?min_risk=60"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="orgee3941d"></a>

### статистика по сайту

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/stats
    {"total":1,"active":1,"blocked":0,"avg_risk":75,"unique_ips":1}
