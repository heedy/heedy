# Minimal makefile for Sphinx documentation
#

# You can set these variables from the command line, and also
# from the environment for the first two.
SPHINXOPTS    ?=
SPHINXBUILD   ?= sphinx-build
SOURCEDIR     = .
BUILDDIR      = _build

# Put it first so that "make" without argument is like "make help".
all: getvars pipesdocs
	@$(SPHINXBUILD) -M html "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)

_govars/extractvar:
	go build -o _govars/extractvar _govars/extractvar.go 

getvars: _govars/extractvar
	_govars/extractvar ../backend/database/create.go schema > _govars/heedy_schema.sql
	_govars/extractvar ../plugins/timeseries/backend/timeseries/sqldata.go sqlSchema > _govars/timeseries_schema.sql
	_govars/extractvar ../plugins/notifications/backend/notifications/sql.go sqlSchema > _govars/notifications_schema.sql
	_govars/extractvar ../plugins/kv/backend/kv/kv.go sqlSchema > _govars/kv_schema.sql

.PHONY: help Makefile

.PHONY: pipesdocs
pipesdocs:
	go get github.com/heedy/pipescript
	go install github.com/heedy/pipescript/pipes
	python3 ./pipescript.py

# Catch-all target: route all unknown targets to Sphinx using the new
# "make mode" option.  $(O) is meant as a shortcut for $(SPHINXOPTS).
%: Makefile getvars pipesdocs
	@$(SPHINXBUILD) -M $@ "$(SOURCEDIR)" "$(BUILDDIR)" $(SPHINXOPTS) $(O)
