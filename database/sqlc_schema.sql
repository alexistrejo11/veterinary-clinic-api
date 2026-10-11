-- SQLC cannot infer types created inside the conditional DO block in 000001.
-- Keep this prelude in sync with that migration; it is not applied at runtime.
CREATE TYPE person_gender AS ENUM ('male', 'female', 'not_specified');
CREATE TYPE veterinarian_speciality AS ENUM (
    'unknown_specialty', 'general_practice', 'surgery', 'internal_medicine', 'dentistry',
    'dermatology', 'oncology', 'cardiology', 'neurology', 'ophthalmology', 'radiology',
    'emergency_critical_care', 'anesthesiology', 'pathology', 'preventive_medicine',
    'exotic_animal_medicine', 'equine_medicine', 'avian_medicine', 'zoo_animal_medicine',
    'food_animal_medicine', 'public_health', 'behavior', 'rehabilitation', 'nutrition'
);
CREATE TYPE currency AS ENUM ('USD', 'MXN', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'SEK', 'NZD');
CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('cash', 'credit_card', 'debit_card', 'bank_transfer', 'paypal', 'stripe', 'check');
CREATE TYPE appointment_status AS ENUM ('pending', 'cancelled', 'confirmed', 'rescheduled', 'completed', 'not_presented');
