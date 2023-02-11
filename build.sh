hash=$(sha256sum main.css | cut -c1-64)
rm ./static/*.css
rm -rf build/
npx tailwindcss -i main.css -o ./static/style."${hash}".css --minify
go build -o ./build/neohealth .
rm ./static/*.css
