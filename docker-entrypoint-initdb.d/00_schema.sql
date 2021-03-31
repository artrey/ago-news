-- табличка с пользователями и их паролями
CREATE TABLE news
(
    id      BIGSERIAL PRIMARY KEY,
    title   VARCHAR(40) NOT NULL,
    text    VARCHAR     NOT NULL,
    image   VARCHAR     NOT NULL,
    created TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);
