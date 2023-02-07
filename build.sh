# shellcheck disable=SC2207
hash=($(sha256sum main.css))
npx tailwindcss -i main.css -o ./static/style."${hash[0]}".css --minify && go build -o ./build/neohealth .
