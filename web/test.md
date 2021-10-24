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

