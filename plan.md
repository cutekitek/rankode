# Rankode LMS Redesign Plan

## Overview
Transform Rankode from a Codewars-like practice platform into a structured Learning Management System (LMS) where teachers can manage courses, assignments, and student progress.

## Phase 1: Integration & Core Backend Enhancements
- [ ] **LTI 1.3 Support**: Implement OIDC flow and Grade/Assignment services for external LMS integration.
- [ ] **Submission-Assignment Linking**: Link `attempts` to `assignments` in the database and models.
- [ ] **Automatic Grading**: Trigger grade creation/updates upon successful task completion.
- [ ] **Real Analytics**: Replace statistics placeholders with actual data queries.

## Phase 2: Frontend Redesign (LMS Module)
- [ ] **Course Management UI**:
    - Teacher: Create/Edit courses, manage roster, join codes.
    - Student: Course list, join via code.
- [ ] **Assignment Workflow**:
    - Teacher: Assignment builder (select tasks, set deadlines, weights).
    - Student: Assignment view with task list, progress, and grades.
- [ ] **Role-Based Navigation**: Distinct views for Teachers and Students.

## Phase 3: Dashboards & Reporting
- [ ] **Teacher Gradebook**: Matrix view of students vs assignments.
- [ ] **Student Dashboard**: "My Progress" and "Upcoming Deadlines" views.

## Implementation Steps (Current)
1.  **Database Update**: Link attempts to assignments.
2.  **API Models**: Update models to reflect the new relationships.
3.  **Frontend Scaffolding**: Create basic Course and Assignment pages.
