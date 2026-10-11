# Veterinary Clinic API

## Problem / Vision

Veterinary clinics need one consistent place to manage client relationships, pet care, scheduling, clinical records, billing records, and notifications. This service provides the shared operational backend for clinic staff and pet owners instead of requiring each workflow to use a separate system.

## Users / Actors

- **Pet owners** manage their profile, addresses, pets, appointments, payments, notifications, and read-only medical history.
- **Receptionists** manage customers, pets, appointments, addresses, and payment records.
- **Veterinarians** manage assigned appointments and clinical records for pets.
- **Administrators** manage users, employees, and clinic-wide operations.

## Domain Model

A **Customer** represents a pet owner and can have a user account. A Customer owns one or more **Pets**. A Pet can have **Appointments** and **Medical Sessions**.

An **Appointment** connects a customer, pet, clinic service, scheduled time, status, and optionally an assigned employee. A **Medical Session** records care for a pet and may be associated with an appointment. A session can contain vaccinations, surgeries, prescriptions, attachments, and billable services.

A **Payment** records a financial transaction associated with clinic work, such as an appointment or medical session. **Notifications** belong to user accounts and are delivered through configured communication channels.

## Business Rules

- A scheduled appointment requires a customer and pet.
- An active appointment cannot overlap another appointment assigned to the same veterinarian.
- No more than five active appointments may be scheduled in the same hour.
- Medical-session extensions belong to a medical session.
- Payments must have a positive amount.
- A paid payment requires a payment date and receipt number.
- Customers can access only their own pets, appointments, payments, notifications, and medical information.

## Scope

**In scope:** identity and access, customer and employee management, pets, addresses, appointment scheduling, medical records, clinical catalogs, payment tracking, and notifications.

**Out of scope:** web or mobile user interfaces, payment-processor settlement, public availability discovery, file storage, and public API hosting.
