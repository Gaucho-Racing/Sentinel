FROM node:22-alpine3.19 as build

WORKDIR /app
COPY . /app

RUN npm install
RUN npm run build

FROM nginx:1.25.2-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]