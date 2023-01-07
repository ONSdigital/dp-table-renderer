FROM onsdigital/dp-concourse-tools-ubuntu-20:ubuntu20.4-rc.1

WORKDIR /app/

COPY ./build/dp-table-renderer .

CMD ./dp-table-renderer
