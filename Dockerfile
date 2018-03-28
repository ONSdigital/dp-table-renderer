FROM onsdigital/dp-concourse-tools-ubuntu

WORKDIR /app/

COPY ./build/dp-table-renderer .

CMD ./dp-table-renderer
