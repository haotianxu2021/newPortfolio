CREATE TABLE "users" (
  "id" SERIAL PRIMARY KEY,
  "username" VARCHAR(50) UNIQUE NOT NULL,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password_hash" VARCHAR(255) NOT NULL,
  "first_name" VARCHAR(50),
  "last_name" VARCHAR(50),
  "bio" TEXT,
  "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE "posts" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER,
  "title" VARCHAR(255) NOT NULL,
  "content" TEXT NOT NULL,
  "type" VARCHAR(50) NOT NULL,
  "status" VARCHAR(20) DEFAULT 'published',
  "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP),
  "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE "images" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER,
  "file_path" VARCHAR(255) NOT NULL,
  "alt_text" VARCHAR(255),
  "uploaded_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP)
);

CREATE TABLE "post_images" (
  "post_id" INTEGER,
  "image_id" INTEGER,
  "display_order" INTEGER,
  PRIMARY KEY ("post_id", "image_id")
);

CREATE TABLE "tags" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE "post_tags" (
  "post_id" INTEGER,
  "tag_id" INTEGER,
  PRIMARY KEY ("post_id", "tag_id")
);

CREATE TABLE "comments" (
  "id" SERIAL PRIMARY KEY,
  "post_id" INTEGER,
  "user_id" INTEGER,
  "content" TEXT NOT NULL,
  "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP)
);

ALTER TABLE "posts" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE;

ALTER TABLE "images" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL;

ALTER TABLE "post_images" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE;

ALTER TABLE "post_images" ADD FOREIGN KEY ("image_id") REFERENCES "images" ("id") ON DELETE CASCADE;

ALTER TABLE "post_tags" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE;

ALTER TABLE "post_tags" ADD FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON DELETE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("post_id") REFERENCES "posts" ("id") ON DELETE CASCADE;

ALTER TABLE "comments" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE SET NULL;
