# Changelog

## pre-3.1

See Git Commit History.

## 3.1

- Routes (collections) can now be dynamically created with the `/meta/` path.
- API now fully documented in `swagger.yaml`
- Authentication OTP secret no longer saved in database.
    - A hidden file `.otp` is created on initialisation.
- CORS host is now required.
    - Configured in `config.toml`
- Added limit parameter for collection item queries.
- `PATCH` method removed from Reference.