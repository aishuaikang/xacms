-- 用户表初始化数据
INSERT INTO
    users (
        id,
        nickname,
        username,
        password,
        email,
        phone,
        avatar,
        role_id,
        created_at,
        updated_at
    )
VALUES (
        1,
        '超级管理员',
        'admin',
        '$2a$10$hIPOrRMri0BmUr3.ICNGA.CKL46kbmONf84RttJOl8zJyNW11YOOe',
        null,
        null,
        null,
        1,
        '2025-09-15 13:16:46.373967+08:00',
        '2025-09-24 10:47:58.412942466+08:00'
    );