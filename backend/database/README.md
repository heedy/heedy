# Database

The database is structured as follows:

There are 4 entities in total:

- groups - groups are an entity that holds streams and apps, as well as scopes. A group's scopes only go up to the owning user's scopes. That is, even if a group has the `user:create` scope, it will not be able to create new users _unless the user owning the group also has this scope_
- users - A user is a group with an additional password, and that can log into the frontend. The owner of the user is itself. A user's scopes encompass the entire database. That is, if a user has the `user:create` scope, it will be permitted to create users.
- apps - A app represents something that has connected to the database programmatically. Apps can represent external programs, such as apps, services and devices, in which case the app will have an API key associated with it, or it can represent a user's instance of a plugin, in which case the app will not have an API key. A app has its own scopes, which work in the same way as group scope, meaning that even if a app has a scope, it will only be permitted to do _up to_ its users' permissions.

# The `heedy` user

When a database is created, the `heedy` user is created automatically. The user has no password,
and nobody can log in as it. The `heedy` user represents heedy, and its plugins.

The heedy user is off-limits, meaning that no matter what your scopes, you cannot add/remove
groups or apps from the user

# Scopes

Scopes are used in groups

- `user:create` - Allows creating new users.
- `user:read` - Allows reading the owner.
- `user:modify` - Allows modifying the user's information
- `users:read` - Allows reading all users (all users that the owner can read), ignoring the group
- `users:modify` - Modify users' information
- `group:create` - Allows creating a group for the user
- `group:
