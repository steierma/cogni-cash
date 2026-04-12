---
name: flutter-entity-mapper
description: Generates and audits manual JSON mapping (fromJson/toJson) for Flutter entities in CogniCash. Use when adding or modifying domain entities in the mobile project.
---

# Flutter Entity Mapper

Enforce stable, manual JSON mapping for Flutter entities to prevent runtime crashes.

## Rules
1. **Manual Mapping Only:** Do NOT use `json_serializable` for entities. Write `fromJson` and `toJson` manually.
2. **Resilient Models:** Use null-safety checks and default values (e.g., `?? 0.0`, `?? ''`) for all fields in `fromJson`.
3. **Date Safety:** Use a helper or try-catch block for `DateTime.parse` in `fromJson`.
4. **Isar Compatibility:** If the entity is a `@collection`, ensure the `isarId` and `_fastHash` logic are present.

## Workflows

### 1. New Entity Creation
- Define the class with `final` fields.
- Implement the constructor with `required` and optional parameters.
- Implement `copyWith`.
- Implement `fromJson` using the resilient pattern:
    ```dart
    factory MyEntity.fromJson(Map<String, dynamic>? json) {
      if (json == null) return MyEntity.empty(); // Define an empty/default factory
      return MyEntity(
        id: (json['id'] ?? '').toString(),
        amount: (json['amount'] as num?)?.toDouble() ?? 0.0,
        // ...
      );
    }
    ```
- Implement `toJson`.

### 2. Audit
When modifying an entity:
1. Ensure every new field is added to `fromJson`, `toJson`, and `copyWith`.
2. Verify that `fromJson` handles `null` values from the API gracefully.

## Automation
- If Isar is used, run `dart run build_runner build --delete-conflicting-outputs` after changes to update `.g.dart` files.
