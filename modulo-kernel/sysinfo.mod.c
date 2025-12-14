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
	{ 0xc7ffe1aa, "si_meminfo" },
	{ 0x058c185a, "jiffies" },
	{ 0xf2c4f3f1, "seq_printf" },
	{ 0x1790826a, "init_task" },
	{ 0x2182515b, "__num_online_cpus" },
	{ 0xbd03ed67, "random_kmalloc_seed" },
	{ 0xfed1e3bc, "kmalloc_caches" },
	{ 0x70db3fe4, "__kmalloc_cache_noprof" },
	{ 0xe8d8d116, "get_task_mm" },
	{ 0xa59da3c0, "down_read" },
	{ 0xa59da3c0, "up_read" },
	{ 0x397daafe, "mmput" },
	{ 0xcb8b6ec6, "kfree" },
	{ 0x04e8afba, "access_process_vm" },
	{ 0xd272d446, "__stack_chk_fail" },
	{ 0x33c78c8a, "remove_proc_entry" },
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
	0xc7ffe1aa,
	0x058c185a,
	0xf2c4f3f1,
	0x1790826a,
	0x2182515b,
	0xbd03ed67,
	0xfed1e3bc,
	0x70db3fe4,
	0xe8d8d116,
	0xa59da3c0,
	0xa59da3c0,
	0x397daafe,
	0xcb8b6ec6,
	0x04e8afba,
	0xd272d446,
	0x33c78c8a,
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
	"si_meminfo\0"
	"jiffies\0"
	"seq_printf\0"
	"init_task\0"
	"__num_online_cpus\0"
	"random_kmalloc_seed\0"
	"kmalloc_caches\0"
	"__kmalloc_cache_noprof\0"
	"get_task_mm\0"
	"down_read\0"
	"up_read\0"
	"mmput\0"
	"kfree\0"
	"access_process_vm\0"
	"__stack_chk_fail\0"
	"remove_proc_entry\0"
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


MODULE_INFO(srcversion, "BC5D62D50F223671176F617");
