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
HAML_DEPS+=$(wildcard frontend/src/templates/*.hbs)
JS_DEPS=$(wildcard frontend/src/js/*.js)
JS_TGTS=$(foreach js,$(JS_DEPS),$(basename $(subst src/js,build,$(js))).min.js)
HBS_DEPS=$(wildcard frontend/src/templates/*.hbs)
HBS_TGTS=$(foreach hbs,$(HBS_DEPS),$(basename $(subst src/templates,build,$(hbs)))_template.min.js)
CSS_DEPS=$(wildcard frontend/src/css/*)
CSS_TGTS=$(foreach css,$(CSS_DEPS),$(basename $(subst src/css,build,$(css))).css)
FRONT_CLEAN=$(HAML_TGTS:%=clean-%)
FRONT_CLEAN+=$(JS_TGTS:%=clean-%)
FRONT_CLEAN+=$(HBS_TGTS:%=clean-%)
FRONT_CLEAN+=$(CSS_TGTS:%=clean-%)

# Whole damn thing
app: backend-test backend frontend

app-clean: backend-clean frontend-clean

pack-all:
	rm -rf /tmp/omapp-pack
	mkdir -p /tmp/omapp-pack/omapp
	mkdir /tmp/omapp-pack/omapp/bin
	mkdir /tmp/omapp-pack/omapp/dist
	find ./cmd -perm +0111 -type f -exec cp '{}' /tmp/omapp-pack/omapp/bin/ \;
	cp ./frontend/dist/* /tmp/omapp-pack/omapp/dist/
	tar -C /tmp/omapp-pack -cJvf omapp-all.tar.xz omapp

pack-frontend:
	rm -rf /tmp/omapp-pack
	mkdir -p /tmp/omapp-pack/omapp
	mkdir /tmp/omapp-pack/omapp/dist
	cp ./frontend/dist/* /tmp/omapp-pack/omapp/dist/
	tar -C /tmp/omapp-pack -cJvf omapp-front.tar.xz omapp

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
	mkdir -p pkg/ver
	echo -e "package ver\nconst VERSION = \"$(VERSION)\"" > pkg/ver/version.go

# Frontend targets
frontend: frontend-clean $(JS_TGTS) $(HBS_TGTS) $(CSS_TGTS) $(HAML_TGTS)
$(JS_TGTS): $(JS_DEPS)
	yuicompressor --type js -o $@ $(subst .min,,$(subst build,src/js,$@))
$(HBS_TGTS): $(HBS_DEPS)
	handlebars -p -m -f $@ $(subst .min.js,.hbs,$(subst build,src/templates,$(subst _template,,$@)))
$(CSS_TGTS): $(CSS_DEPS)
	yuicompressor --type css -o $@ $(subst build,src/css,$@)
$(HAML_TGTS): $(HAML_DEPS)
	haml --trace -r ./frontend/src/helpers.rb $(subst .html,.haml,$(subst dist,src,$@)) > $@

frontend-clean: $(FRONT_CLEAN)
$(FRONT_CLEAN):
	rm -f $(@:clean-%=%)

# useful for debug
print-%:
	@echo $* = $($*)

.PHONY: app-clean backend-test backend-clean frontend frontend-clean
.PHONY: pack-all pack-frontend
