-- PostgreSQL 样例数据

-- 插入测试用户
INSERT INTO users (username, email, password_hash, full_name, role) VALUES
    ('admin', 'admin@example.com', crypt('admin123', gen_salt('bf')), '系统管理员', 'admin'),
    ('teacher_zhang', 'zhang@example.com', crypt('teacher123', gen_salt('bf')), '张老师', 'teacher'),
    ('teacher_wang', 'wang@example.com', crypt('teacher123', gen_salt('bf')), '王老师', 'teacher');

-- 插入样例教案
INSERT INTO lessons (user_id, title, subject, grade, topic, duration, content, objectives, key_points, difficult_points, status) 
VALUES (
    (SELECT id FROM users WHERE username = 'teacher_zhang'),
    '一元一次方程',
    '数学',
    '七年级',
    '一元一次方程的解法',
    45,
    '{
        "sections": [
            {
                "title": "导入新课",
                "duration": 5,
                "teacher_activity": "通过生活实例引入方程的概念",
                "student_activity": "观察思考，回答问题"
            },
            {
                "title": "新课讲解",
                "duration": 25,
                "teacher_activity": "讲解一元一次方程的定义、标准形式和解法",
                "student_activity": "认真听讲，做笔记，提出疑问"
            },
            {
                "title": "课堂练习",
                "duration": 10,
                "teacher_activity": "巡视指导，个别辅导",
                "student_activity": "独立完成练习题"
            },
            {
                "title": "总结归纳",
                "duration": 5,
                "teacher_activity": "总结本节课重点内容",
                "student_activity": "回顾总结，提出问题"
            }
        ],
        "materials": ["多媒体课件", "练习题", "白板"],
        "homework": "完成课后练习1-10题"
    }'::jsonb,
    '{
        "knowledge": "掌握一元一次方程的定义和标准形式，学会解一元一次方程",
        "process": "通过观察、分析、归纳，培养学生的逻辑思维能力",
        "emotion": "感受数学在生活中的应用，培养学习兴趣"
    }'::jsonb,
    '["一元一次方程的定义", "一元一次方程的解法", "移项法则"]'::jsonb,
    '["方程的等价变形", "符号问题的处理"]'::jsonb,
    'published'
),
(
    (SELECT id FROM users WHERE username = 'teacher_wang'),
    '古诗鉴赏：静夜思',
    '语文',
    '三年级',
    '唐诗鉴赏',
    40,
    '{
        "sections": [
            {
                "title": "诗歌导入",
                "duration": 5,
                "content": "播放古诗朗诵，创设情境"
            },
            {
                "title": "诗歌解读",
                "duration": 20,
                "content": "逐句讲解诗意，分析意象"
            },
            {
                "title": "情感体会",
                "duration": 10,
                "content": "体会诗人思乡之情"
            },
            {
                "title": "诵读表演",
                "duration": 5,
                "content": "学生朗诵展示"
            }
        ]
    }'::jsonb,
    '{
        "knowledge": "理解《静夜思》的诗意，背诵全诗",
        "process": "学习诗歌鉴赏的基本方法",
        "emotion": "体会诗人的思乡之情，培养爱国情怀"
    }'::jsonb,
    '["诗歌朗读", "诗意理解", "情感体悟"]'::jsonb,
    '["意象的理解", "情感的把握"]'::jsonb,
    'draft'
);

SELECT 'Sample data inserted successfully!' AS status;
