docker compose \
-f docker-compose-prod.yml \
config > docker-stack.yml

sed -i '/^\s*name:\s*[a-z]/d' docker-stack.yml

# sed -i 1d docker-stack.yml
(echo 'version: "3.3"' | cat - docker-stack.yml) > docker-stack.yml.tmp && mv docker-stack.yml.tmp docker-stack.yml