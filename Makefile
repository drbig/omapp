# Configuration below
CMD_LIST=backend model uploader worker
PKG_LIST=model overmapper queue web

# Automated discovery of what depends on what
PKG_TESTS=$(PKG_LIST:%=test-%)
PKG_DEPS=$(foreach pkg,$(PKG_LIST),$(wildcard pkg/$(pkg)/*.go))
CMD_TGTS=$(foreach cmd,$(CMD_LIST),cmd/$(cmd)/$(cmd))
CMD_DEPS=$(foreach cmd,$(CMD_LIST),$(wildcard cmd/$(cmd)/*.go))
CLEAN_TGTS=$(CMD_TGTS:%=clean-%)
VERSION=$(shell git describe --tags --always --dirty)
HAML_MAIN=$(wildcard frontend/src/*.haml)
HAML_TGTS=$(foreach temp,$(HAML_MAIN),$(basename $(subst src,dist,$(temp))).html)
HAML_DEPS=$(HAML_MAIN)
HAML_DEPS+=$(wildcard frontend/src/partials/*.haml)
HAML_DEPS+=$(wildcard frontend/src/js/*.js)
HAML_CLEAN=$(HAML_TGTS:%=clean-%)
JS_DEPS=$(wildcard frontend/src/js/*.js)
JS_TGTS=$(foreach js,$(JS_DEPS),$(js).min)
JS_CLEAN=$(JS_TGTS:%=clean-%)

# Backend targets
backend: version $(CMD_TGTS)
$(CMD_TGTS): $(CMD_DEPS) $(PKG_DEPS)
	cd $(dir $@) && go build

backend-test: $(PKG_TESTS)
$(PKG_TESTS):
	cd pkg/$(@:test-%=%) && go test

backend-clean: $(CLEAN_TGTS)
$(CLEAN_TGTS):
	rm -f $(@:clean-%=%)

version:
	echo -e "package ver\nconst VERSION = \"$(VERSION)\"" > pkg/ver/version.go

# Frontend targets
frontend: $(JS_TGTS) $(HAML_TGTS)
$(HAML_TGTS): $(HAML_DEPS)
	haml -r ./frontend/src/helpers.rb $< > $@
$(JS_TGTS): $(JS_DEPS)
	yuicompressor --type js $<

frontend-clean: $(HAML_CLEAN) $(JS_CLEAN)
$(HAML_CLEAN):
	rm -f $(@:clean-%=%)
$(JS_CLEAN):
	rm -f $(@:clean-%=%)

# useful for debug
print-%:
	@echo $* = $($*)

.PHONY: test clean
