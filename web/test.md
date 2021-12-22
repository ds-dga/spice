curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username":"xyz","password":"xyz"}' \
  http://localhost:3300/hook/uptime


curl \
  --request POST \
  --data '{"username":"xyz","password":"xyz"}' \
  http://localhost:3300/hook/uptime


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url":"https://example.com/console/data/arrakis/schema/public/tables/api/modify","password":"xyz"}' \
  http://localhost:3300/hook/uptime


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url":"https://example.com/console/data/arrakis/schema/public/tables/api/modify","status_code":200}' \
  http://localhost:3300/hook/uptime


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url":"https://example.com/console/data/arrakis/schema/public/tables/api/modify","status_code":200,"response_time_ms":53.33,"size_byte":"33321"}' \
  http://localhost:3300/hook/uptime

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"url":"https://example.com/console/data/arrakis/schema/public/tables/api/modify","status_code":200,"response_time_ms":53.33,"size_byte":33321,"from":"10coffee","from_coords":[133,10]}' \
  http://localhost:3300/hook/uptime


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"id":"27d7e0ba-e632-46b5-85ed-7ec46eaf8b2a","status_code":400,"response_time_ms":53.33,"size_byte":33321,"from":"10coffee","from_coords":[133,10]}' \
  http://localhost:3300/hook/uptime

curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"id":"27d7e0ba-e632-46b5-85ed-7ec46eaf8b2a","status_code":400,"response_time_ms":563.33,"size_byte":33321,"from":"10coffee","from_coords":[133,10]}' \
  https://ds.10z.dev/hook/uptime


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"id":"27d7e0ba-e632-46b5-85ed-7ec46eaf8b2a","status_code":200,"response_time_ms":83.33,"size_byte":33321,"from":"10coffee","from_coords":[133.2,10]}' \
  https://ds.10z.dev/hook/uptime


curl --header "Content-Type: application/json" --request POST --data '{"url":"https://showtimes.everyday.in.th","name":"stth-dj","group":"sipp11","frequency":"hourly","extras":{"web":"https://i.imgur.com/bQ0BnqY.png","resource_id":"resource xyz","package_id":"package abc"}}' http://10.1.1.50:3300/uptime


curl --header "Content-Type: application/json" --request POST --data '{"url":"https://test.10ninox.com","name":"10ninox-test","group":"sipp11","frequency":"never"}' http://10.1.1.50:3300/uptime


---------------------------

# Auth

Register


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"email":"sipp11@duck.com","password":"whatever.is.secure","first_name":"Sipp","last_name":"Ninox"}' \
  http://127.0.0.1:3300/signup


curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"email":"sipp11@duck.com","password":"whatever.is.secure"}' \
  http://127.0.0.1:3300/login
