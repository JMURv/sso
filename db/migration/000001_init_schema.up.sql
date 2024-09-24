-- users table
CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    name       VARCHAR(50)  NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    email      VARCHAR(50)  NOT NULL UNIQUE,

    avatar     VARCHAR(255),
    address    VARCHAR(255),
    phone      VARCHAR(20),
    is_opt BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- carts table
CREATE TABLE IF NOT EXISTS carts
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    user_id    UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- categories
CREATE TABLE IF NOT EXISTS categories
(
    slug             VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,

    title            VARCHAR(255) UNIQUE             NOT NULL,
    product_quantity INTEGER,

    src              VARCHAR(255),
    alt              VARCHAR(255),

    parent_slug      VARCHAR(255),
    FOREIGN KEY (parent_slug) REFERENCES categories (slug) ON DELETE SET NULL
);

-- items
CREATE TABLE IF NOT EXISTS items
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    price         DECIMAL(10, 2),
    in_stock      BOOLEAN          DEFAULT true,

    item_type     VARCHAR(50),
    src           VARCHAR(255),
    alt           VARCHAR(255),

    category_slug VARCHAR(255),
    parent_id     UUID, -- self-referencing foreign key for variants
    FOREIGN KEY (category_slug) REFERENCES categories (slug) ON DELETE SET NULL,
    FOREIGN KEY (parent_id) REFERENCES items (id) ON DELETE CASCADE
);

-- orders
CREATE TABLE IF NOT EXISTS cart_items
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    quantity   INTEGER NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    cart_id    UUID    NOT NULL,
    product_id UUID    NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES carts (id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS orders
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    status       VARCHAR(50)    NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,

    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at   TIMESTAMP WITH TIME ZONE,

    user_id      UUID           NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS order_items
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    quantity   INTEGER        NOT NULL,
    price      DECIMAL(10, 2) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    order_id   UUID           NOT NULL,
    item_id    UUID           NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS item_media
(
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    src     VARCHAR(255) NOT NULL,
    alt     VARCHAR(255) NOT NULL,

    item_id UUID,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS item_attrs
(
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name    VARCHAR(255) NOT NULL,
    value   VARCHAR(255) NOT NULL,

    item_id UUID,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS related_products
(
    PRIMARY KEY (item_id, related_item_id),
    item_id         UUID,
    related_item_id UUID,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
    FOREIGN KEY (related_item_id) REFERENCES items (id) ON DELETE CASCADE
);

-- favorites
CREATE TABLE IF NOT EXISTS favorites
(
    user_id UUID NOT NULL,
    item_id UUID NOT NULL,
    PRIMARY KEY (user_id, item_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

-- blog
CREATE TABLE IF NOT EXISTS blog_posts
(
    slug        VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL,
    title       VARCHAR(255)                    NOT NULL,
    description TEXT,
    src         VARCHAR(255),
    alt         VARCHAR(255),

    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS blog_media
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    src       VARCHAR(255) NOT NULL,
    alt       VARCHAR(255) NOT NULL,

    blog_slug VARCHAR(255),
    FOREIGN KEY (blog_slug) REFERENCES blog_posts (slug) ON DELETE CASCADE
);

-- promotion
CREATE TABLE IF NOT EXISTS promotions
(
    slug        VARCHAR(255) PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    src         VARCHAR(255) NOT NULL,
    alt         VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS promotion_items
(
    promotion_slug VARCHAR(255),
    item_id        UUID,
    PRIMARY KEY (promotion_slug, item_id),
    FOREIGN KEY (promotion_slug) REFERENCES promotions (slug) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE
);

-- pages
CREATE TABLE IF NOT EXISTS pages
(
    slug VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255),
    href VARCHAR(255)
);

-- SEO
CREATE TABLE IF NOT EXISTS seo
(
    id             serial PRIMARY KEY,
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    keywords       TEXT,
    OGTitle        VARCHAR(255),
    OGDescription  TEXT,

    page_slug      VARCHAR(255),
    FOREIGN KEY (page_slug) REFERENCES pages (slug) ON DELETE CASCADE,

    promotion_slug VARCHAR(255),
    FOREIGN KEY (promotion_slug) REFERENCES promotions (slug) ON DELETE CASCADE,

    product_id     UUID,
    FOREIGN KEY (product_id) REFERENCES items (id) ON DELETE CASCADE,

    category_slug  VARCHAR(255),
    FOREIGN KEY (category_slug) REFERENCES categories (slug) ON DELETE CASCADE

);

-- banner
CREATE TABLE IF NOT EXISTS banner
(
    id          serial PRIMARY KEY,

    page_slug  VARCHAR(255),
    FOREIGN KEY (page_slug) REFERENCES pages (slug) ON DELETE SET NULL,

    promotion_slug VARCHAR(255),
    FOREIGN KEY (promotion_slug) REFERENCES promotions (slug) ON DELETE CASCADE,

    category_slug  VARCHAR(255),
    FOREIGN KEY (category_slug) REFERENCES categories (slug) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS slides
(
    id          serial PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    href        VARCHAR(255) NOT NULL,
    scr         VARCHAR(255) NOT NULL,
    alt         VARCHAR(255),
    buttonText  VARCHAR(255),
    buttonHref  VARCHAR(255),

    banner_id   INT,
    FOREIGN KEY (banner_id) REFERENCES banner (id) ON DELETE CASCADE
);

-- partners
CREATE TABLE IF NOT EXISTS partners
(
    id    serial PRIMARY KEY,
    title VARCHAR(255) PRIMARY KEY,
    href  VARCHAR(255),
    scr   VARCHAR(255)
);

-- reviews
CREATE TABLE IF NOT EXISTS reviews
(
    id          serial PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL
);