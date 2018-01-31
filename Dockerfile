FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-table-renderer .

CMD ./dp-table-renderer
