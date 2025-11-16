-- Удаляем таблицы в обратном порядке (из-за foreign keys)
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- Удаляем расширение UUID (опционально, может использоваться другими БД)
DROP EXTENSION IF EXISTS "uuid-ossp";
