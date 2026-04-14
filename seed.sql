INSERT INTO
    quiz_packages (
        id,
        title,
        description,
        category,
        difficulty_level,
        thumbnail_url,
        is_free,
        price,
        currency,
        total_questions,
        duration_minutes,
        passing_score,
        max_attempts,
        total_taken,
        average_score,
        is_published,
        published_at,
        created_at
    )
VALUES (
        '660e8400-e29b-41d4-a716-446655440001',
        'Basic Nutrition Quiz',
        'Learn about nutrition basics',
        'Nutrition',
        'easy',
        'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?w=400&h=300&fit=crop',
        true,
        0,
        'IDR',
        5,
        15,
        60,
        5,
        120,
        75.5,
        true,
        NOW(),
        NOW()
    ),
    (
        '660e8400-e29b-41d4-a716-446655440002',
        'Fitness Fundamentals',
        'Basic fitness knowledge',
        'Fitness',
        'easy',
        'https://picsum.photos/500/300',
        true,
        0,
        'IDR',
        5,
        20,
        65,
        3,
        95,
        72.3,
        true,
        NOW(),
        NOW()
    ),
    (
        '660e8400-e29b-41d4-a716-446655440003',
        'Advanced Nutrition',
        'Deep nutrition concepts',
        'Nutrition',
        'medium',
        'https://fastly.picsum.photos/id/935/500/300.jpg?hmac=Y8t_zs_-ZWg1u6xYxdJh_FnC4IUFJoGlThPdXIO9XJ8',
        false,
        50000,
        'IDR',
        5,
        25,
        70,
        3,
        40,
        68.5,
        true,
        NOW(),
        NOW()
    ),
    (
        '660e8400-e29b-41d4-a716-446655440004',
        'Workout Science',
        'Exercise physiology',
        'Fitness',
        'medium',
        'https://fastly.picsum.photos/id/341/500/300.jpg?hmac=2xCwco9WhzoamRSJQfg0QoBU6Q0IqUBZ-0FGmeevn_I',
        false,
        60000,
        'IDR',
        5,
        25,
        70,
        3,
        30,
        66.2,
        true,
        NOW(),
        NOW()
    );

INSERT INTO
    quiz_questions (
        id,
        quiz_package_id,
        question_text,
        question_image_url,
        question_order,
        question_type,
        number_of_options,
        explanation,
        point,
        created_at
    )
VALUES

-- PACKAGE 1
(
    '770e8400-e29b-41d4-a716-446655440001',
    '660e8400-e29b-41d4-a716-446655440001',
    'Which macronutrient builds muscle?',
    'https://api.example.com/questions/protein.jpg',
    1,
    'multiple_choice',
    4,
    'Proteins build muscle',
    10,
    NOW()
),
(
    '770e8400-e29b-41d4-a716-446655440002',
    '660e8400-e29b-41d4-a716-446655440001',
    'Recommended calorie intake?',
    NULL,
    2,
    'multiple_choice',
    4,
    '2000-2500 kcal',
    10,
    NOW()
),

-- PACKAGE 2
(
    '770e8400-e29b-41d4-a716-446655440003',
    '660e8400-e29b-41d4-a716-446655440002',
    'Exercise per day?',
    NULL,
    1,
    'multiple_choice',
    4,
    '30 minutes',
    10,
    NOW()
),
(
    '770e8400-e29b-41d4-a716-446655440004',
    '660e8400-e29b-41d4-a716-446655440002',
    'Best exercise for heart?',
    NULL,
    2,
    'multiple_choice',
    4,
    'Cardio',
    10,
    NOW()
),

-- PACKAGE 3
(
    '770e8400-e29b-41d4-a716-446655440005',
    '660e8400-e29b-41d4-a716-446655440003',
    'Vitamin C source?',
    NULL,
    1,
    'multiple_choice',
    4,
    'Orange',
    10,
    NOW()
),
(
    '770e8400-e29b-41d4-a716-446655440006',
    '660e8400-e29b-41d4-a716-446655440003',
    'Protein source?',
    NULL,
    2,
    'multiple_choice',
    4,
    'Egg',
    10,
    NOW()
),

-- PACKAGE 4
(
    '770e8400-e29b-41d4-a716-446655440007',
    '660e8400-e29b-41d4-a716-446655440004',
    'Muscle growth requires?',
    NULL,
    1,
    'multiple_choice',
    4,
    'Training + protein',
    10,
    NOW()
),
(
    '770e8400-e29b-41d4-a716-446655440008',
    '660e8400-e29b-41d4-a716-446655440004',
    'Cardio improves?',
    NULL,
    2,
    'multiple_choice',
    4,
    'Heart health',
    10,
    NOW()
);

INSERT INTO
    quiz_options (
        id,
        quiz_question_id,
        option_text,
        option_image_url,
        option_order,
        is_correct,
        created_at
    )
VALUES

-- Q1
(
    '880e8400-e29b-41d4-a716-446655440001',
    '770e8400-e29b-41d4-a716-446655440001',
    'Carbohydrates',
    NULL,
    1,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440002',
    '770e8400-e29b-41d4-a716-446655440001',
    'Proteins',
    NULL,
    2,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440003',
    '770e8400-e29b-41d4-a716-446655440001',
    'Fats',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440004',
    '770e8400-e29b-41d4-a716-446655440001',
    'Fiber',
    NULL,
    4,
    false,
    NOW()
),

-- Q2
(
    '880e8400-e29b-41d4-a716-446655440005',
    '770e8400-e29b-41d4-a716-446655440002',
    '1500-2000',
    NULL,
    1,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440006',
    '770e8400-e29b-41d4-a716-446655440002',
    '2000-2500',
    NULL,
    2,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440007',
    '770e8400-e29b-41d4-a716-446655440002',
    '3000+',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440008',
    '770e8400-e29b-41d4-a716-446655440002',
    '4000+',
    NULL,
    4,
    false,
    NOW()
),

-- Q3
(
    '880e8400-e29b-41d4-a716-446655440009',
    '770e8400-e29b-41d4-a716-446655440003',
    '10 min',
    NULL,
    1,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440010',
    '770e8400-e29b-41d4-a716-446655440003',
    '20 min',
    NULL,
    2,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440011',
    '770e8400-e29b-41d4-a716-446655440003',
    '30 min',
    NULL,
    3,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440012',
    '770e8400-e29b-41d4-a716-446655440003',
    '60 min',
    NULL,
    4,
    false,
    NOW()
),

-- Q4
(
    '880e8400-e29b-41d4-a716-446655440013',
    '770e8400-e29b-41d4-a716-446655440004',
    'Weightlifting',
    NULL,
    1,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440014',
    '770e8400-e29b-41d4-a716-446655440004',
    'Cardio',
    NULL,
    2,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440015',
    '770e8400-e29b-41d4-a716-446655440004',
    'Yoga',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440016',
    '770e8400-e29b-41d4-a716-446655440004',
    'Stretching',
    NULL,
    4,
    false,
    NOW()
),

-- Q5
(
    '880e8400-e29b-41d4-a716-446655440017',
    '770e8400-e29b-41d4-a716-446655440005',
    'Orange',
    NULL,
    1,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440018',
    '770e8400-e29b-41d4-a716-446655440005',
    'Rice',
    NULL,
    2,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440019',
    '770e8400-e29b-41d4-a716-446655440005',
    'Bread',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440020',
    '770e8400-e29b-41d4-a716-446655440005',
    'Oil',
    NULL,
    4,
    false,
    NOW()
),

-- Q6
(
    '880e8400-e29b-41d4-a716-446655440021',
    '770e8400-e29b-41d4-a716-446655440006',
    'Egg',
    NULL,
    1,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440022',
    '770e8400-e29b-41d4-a716-446655440006',
    'Sugar',
    NULL,
    2,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440023',
    '770e8400-e29b-41d4-a716-446655440006',
    'Salt',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440024',
    '770e8400-e29b-41d4-a716-446655440006',
    'Water',
    NULL,
    4,
    false,
    NOW()
),

-- Q7
(
    '880e8400-e29b-41d4-a716-446655440025',
    '770e8400-e29b-41d4-a716-446655440007',
    'Training + protein',
    NULL,
    1,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440026',
    '770e8400-e29b-41d4-a716-446655440007',
    'Sleep only',
    NULL,
    2,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440027',
    '770e8400-e29b-41d4-a716-446655440007',
    'Water only',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440028',
    '770e8400-e29b-41d4-a716-446655440007',
    'Stretching',
    NULL,
    4,
    false,
    NOW()
),

-- Q8
(
    '880e8400-e29b-41d4-a716-446655440029',
    '770e8400-e29b-41d4-a716-446655440008',
    'Heart health',
    NULL,
    1,
    true,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440030',
    '770e8400-e29b-41d4-a716-446655440008',
    'Muscle only',
    NULL,
    2,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440031',
    '770e8400-e29b-41d4-a716-446655440008',
    'Flexibility',
    NULL,
    3,
    false,
    NOW()
),
(
    '880e8400-e29b-41d4-a716-446655440032',
    '770e8400-e29b-41d4-a716-446655440008',
    'Balance',
    NULL,
    4,
    false,
    NOW()
);