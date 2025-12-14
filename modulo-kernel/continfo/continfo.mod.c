#include <linux/module.h>
#include <linux/export-internal.h>
#include <linux/compiler.h>

MODULE_INFO(name, KBUILD_MODNAME);

__visible struct module __this_module
__section(".gnu.linkonce.this_module") = {
	.name = KBUILD_MODNAME,
	.init = init_module,
#ifdef CONFIG_MODULE_UNLOAD
	.exit = cleanup_module,
#endif
	.arch = MODULE_ARCH_INIT,
};



static const struct modversion_info ____versions[]
__used __section("__versions") = {
	{ 0x003b23f9, "single_open" },
	{ 0xcb8b6ec6, "kfree" },
	{ 0x90a48d82, "__ubsan_handle_out_of_bounds" },
	{ 0xa482e95f, "seq_putc" },
	{ 0x08fc6844, "seq_write" },
	{ 0xf2c4f3f1, "seq_printf" },
	{ 0x33c78c8a, "remove_proc_entry" },
	{ 0x1790826a, "init_task" },
	{ 0xa7ab721e, "kernfs_path_from_node" },
	{ 0x17545440, "strstr" },
	{ 0x40a621c5, "snprintf" },
	{ 0x2435d559, "strncmp" },
	{ 0xbd03ed67, "random_kmalloc_seed" },
	{ 0xfed1e3bc, "kmalloc_caches" },
	{ 0x70db3fe4, "__kmalloc_cache_noprof" },
	{ 0x9479a1e8, "strnlen" },
	{ 0xd70733be, "sized_strscpy" },
	{ 0xe54e0a6b, "__fortify_panic" },
	{ 0xc609ff70, "strncpy" },
	{ 0xd272d446, "__stack_chk_fail" },
	{ 0xbd4e501f, "seq_read" },
	{ 0xfc8fa4ce, "seq_lseek" },
	{ 0xcb077514, "single_release" },
	{ 0xd272d446, "__fentry__" },
	{ 0x82c6f73b, "proc_create" },
	{ 0xe8213e80, "_printk" },
	{ 0xd272d446, "__x86_return_thunk" },
	{ 0xba157484, "module_layout" },
};

static const u32 ____version_ext_crcs[]
__used __section("__version_ext_crcs") = {
	0x003b23f9,
	0xcb8b6ec6,
	0x90a48d82,
	0xa482e95f,
	0x08fc6844,
	0xf2c4f3f1,
	0x33c78c8a,
	0x1790826a,
	0xa7ab721e,
	0x17545440,
	0x40a621c5,
	0x2435d559,
	0xbd03ed67,
	0xfed1e3bc,
	0x70db3fe4,
	0x9479a1e8,
	0xd70733be,
	0xe54e0a6b,
	0xc609ff70,
	0xd272d446,
	0xbd4e501f,
	0xfc8fa4ce,
	0xcb077514,
	0xd272d446,
	0x82c6f73b,
	0xe8213e80,
	0xd272d446,
	0xba157484,
};
static const char ____version_ext_names[]
__used __section("__version_ext_names") =
	"single_open\0"
	"kfree\0"
	"__ubsan_handle_out_of_bounds\0"
	"seq_putc\0"
	"seq_write\0"
	"seq_printf\0"
	"remove_proc_entry\0"
	"init_task\0"
	"kernfs_path_from_node\0"
	"strstr\0"
	"snprintf\0"
	"strncmp\0"
	"random_kmalloc_seed\0"
	"kmalloc_caches\0"
	"__kmalloc_cache_noprof\0"
	"strnlen\0"
	"sized_strscpy\0"
	"__fortify_panic\0"
	"strncpy\0"
	"__stack_chk_fail\0"
	"seq_read\0"
	"seq_lseek\0"
	"single_release\0"
	"__fentry__\0"
	"proc_create\0"
	"_printk\0"
	"__x86_return_thunk\0"
	"module_layout\0"
;

MODULE_INFO(depends, "");


MODULE_INFO(srcversion, "56416653587458FA6FED97A");
