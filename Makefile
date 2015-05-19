CMD_LIST=backend model uploader worker
PKG_LIST=model overmapper queue web

PKG_TESTS=$(PKG_LIST:%=test-%)
PKG_DEPS=$(foreach pkg,$(PKG_LIST),$(wildcard pkg/$(pkg)/*.go))
CMD_TGTS=$(foreach cmd,$(CMD_LIST),cmd/$(cmd)/$(cmd))
CMD_DEPS=$(foreach cmd,$(CMD_LIST),$(wildcard cmd/$(cmd)/*.go))
CLEAN_TGTS=$(CMD_TGTS:%=clean-%)

all: $(CMD_TGTS)
$(CMD_TGTS): $(CMD_DEPS) $(PKG_DEPS)
	cd $(dir $@) && go build

test: $(PKG_TESTS)
$(PKG_TESTS):
	cd pkg/$(@:test-%=%) && go test

clean: $(CLEAN_TGTS)
$(CLEAN_TGTS):
	rm -f $(@:clean-%=%)

.PHONY: test clean

# useful for debug
print-%:
	@echo $* = $($*)
