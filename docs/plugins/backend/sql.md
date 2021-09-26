# Database Schema

Heedy uses an sqlite database, which is located at `data/heedy.db` in the heedy database folder. Any plugins that access or modify the database should have sqlite's foreign keys on, and be compiled with the `json1` extension.

## Core Schema

```{eval-rst}
.. literalinclude:: ../../_govars/heedy_schema.sql
    :language: sql
```

## Timeseries

```{eval-rst}
.. literalinclude:: ../../_govars/timeseries_schema.sql
    :language: sql
```

## Dashboards

```{eval-rst}
.. literalinclude:: ../../_govars/dashboard_schema.sql
    :language: sql
```

## Notifications

```{eval-rst}
.. literalinclude:: ../../_govars/notifications_schema.sql
    :language: sql
```

## Key-Value Storage

```{eval-rst}
.. literalinclude:: ../../_govars/kv_schema.sql
    :language: sql
```
