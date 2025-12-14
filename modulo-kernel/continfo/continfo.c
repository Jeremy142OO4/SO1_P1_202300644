#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/init.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/sched/signal.h>
#include <linux/mm.h>
#include <linux/hashtable.h>
#include <linux/slab.h>
#include <linux/cgroup.h>
#include <linux/string.h>

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Jeremy");
MODULE_DESCRIPTION("Modulo para leer info de contenedores (cgroups) en JSON");
MODULE_VERSION("1.0");

#define PROC_NAME "continfo_so1_202300644"

#define MAX_CGROUP_PATH 256
#define MAX_ID_LEN      64


static void seq_print_json_string(struct seq_file *m, const char *s)
{
    const unsigned char *p = (const unsigned char *)s;
    seq_putc(m, '"');

    while (*p) {
        switch (*p) {
            case '\"': seq_puts(m, "\\\""); break;
            case '\\': seq_puts(m, "\\\\"); break;
            case '\b': seq_puts(m, "\\b");  break;
            case '\f': seq_puts(m, "\\f");  break;
            case '\n': seq_puts(m, "\\n");  break;
            case '\r': seq_puts(m, "\\r");  break;
            case '\t': seq_puts(m, "\\t");  break;
            default:
                if (*p < 0x20)
                    seq_printf(m, "\\u%04x", *p);
                else
                    seq_putc(m, *p);
        }
        p++;
    }
    seq_putc(m, '"');
}

/* ---- Helpers para extraer un ID hex de 64 chars del cgroup path ---- */
static inline bool is_hex(char c)
{
    return (c >= '0' && c <= '9') ||
           (c >= 'a' && c <= 'f') ||
           (c >= 'A' && c <= 'F');
}


static bool extract_container_hex_id(const char *path, char out[MAX_ID_LEN + 1])
{
    int i = 0;
    int run = 0;
    int start = -1;

    if (!path) return false;

    for (i = 0; path[i] != '\0'; i++) {
        if (is_hex(path[i])) {
            if (run == 0) start = i;
            run++;
            if (run == MAX_ID_LEN) {
                memcpy(out, &path[start], MAX_ID_LEN);
                out[MAX_ID_LEN] = '\0';
                return true;
            }
        } else {
            run = 0;
            start = -1;
        }
    }
    return false;
}

/* ---- Estructura para acumulación ---- */
struct cont_agg {
    char id[MAX_ID_LEN + 1];     /* container_id (hex 64) o "unknown" */
    char cgroup_path[MAX_CGROUP_PATH];

    u64 rss_kb;
    u64 cpu_jiffies;
    u32 procs;

    struct hlist_node node;
};

DEFINE_HASHTABLE(cont_table, 10); /* 1024 buckets */

/* hash key: simple de los primeros bytes del id */
static u32 id_hash(const char *id)
{
    /* djb2 */
    u32 h = 5381;
    int c;
    while ((c = *id++))
        h = ((h << 5) + h) + (u32)c;
    return h;
}

static struct cont_agg *find_or_create(const char *id, const char *cpath)
{
    struct cont_agg *e;
    u32 key = id_hash(id);

    hash_for_each_possible(cont_table, e, node, key) {
        if (strncmp(e->id, id, MAX_ID_LEN) == 0)
            return e;
    }

    e = kzalloc(sizeof(*e), GFP_KERNEL);
    if (!e) return NULL;

    strscpy(e->id, id, sizeof(e->id));
    strscpy(e->cgroup_path, cpath ? cpath : "N/A", sizeof(e->cgroup_path));
    e->rss_kb = 0;
    e->cpu_jiffies = 0;
    e->procs = 0;

    hash_add(cont_table, &e->node, key);
    return e;
}

static void free_table(void)
{
    int bkt;
    struct cont_agg *e;
    struct hlist_node *tmp;

    hash_for_each_safe(cont_table, bkt, tmp, e, node) {
        hash_del(&e->node);
        kfree(e);
    }
}

/* ---- show ---- */
static int continfo_show(struct seq_file *m, void *v)
{
    struct task_struct *task;
    int bkt;
    struct cont_agg *e;
    bool first = true;
    u32 count = 0;

    /* limpiar tabla por lectura (recalcular fresh) */
    free_table();

    /* recorrer procesos y agrupar por cgroup */
    for_each_process(task) {
        char path[MAX_CGROUP_PATH] = {0};
        char id[MAX_ID_LEN + 1] = {0};
        struct cgroup *cg;
        u64 rss_kb = 0;
        u64 cpu_j = 0;

        /* RSS: solo si tiene mm */
        if (task->mm) {
            rss_kb = (u64)get_mm_rss(task->mm) << (PAGE_SHIFT - 10);
        }

        /* CPU: jiffies (utime+stime) */
        cpu_j = (u64)task->utime + (u64)task->stime;

        /* cgroup path (v2 default cgroup) */
        cg = task_dfl_cgroup(task);
        if (cg) {
            /* cgroup_path escribe el path completo */
            cgroup_path(cg, path, sizeof(path));
        } else {
            strncpy(path, "unknown", sizeof(path));
        }


        if (!(strstr(path, "docker") || strstr(path, "containerd") || strstr(path, "kubepods"))) {
            continue;
        }

        if (!extract_container_hex_id(path, id)) {

            snprintf(id, sizeof(id), "path:%u", id_hash(path));
        }

        e = find_or_create(id, path);
        if (!e) continue;

        e->rss_kb += rss_kb;
        e->cpu_jiffies += cpu_j;
        e->procs += 1;
    }

    /* contar entradas */
    hash_for_each(cont_table, bkt, e, node) {
        count++;
    }

    /* JSON */
    seq_printf(m, "{\n");
    seq_printf(m, "  \"Count\": %u,\n", count);
    seq_printf(m, "  \"Containers\": [\n");

    hash_for_each(cont_table, bkt, e, node) {
        if (!first) seq_printf(m, ",\n");
        first = false;

        seq_printf(m, "    {\n");
        seq_puts(m, "      \"ContainerID\": ");
        seq_print_json_string(m, e->id);
        seq_puts(m, ",\n");

        seq_puts(m, "      \"CgroupPath\": ");
        seq_print_json_string(m, e->cgroup_path);
        seq_puts(m, ",\n");

        seq_printf(m, "      \"RSS_KB\": %llu,\n", (unsigned long long)e->rss_kb);
        seq_printf(m, "      \"CPU_Jiffies\": %llu,\n", (unsigned long long)e->cpu_jiffies);
        seq_printf(m, "      \"Procs\": %u\n", e->procs);
        seq_printf(m, "    }");
    }

    seq_printf(m, "\n  ]\n}\n");
    return 0;
}

static int continfo_open(struct inode *inode, struct file *file)
{
    return single_open(file, continfo_show, NULL);
}

static const struct proc_ops continfo_ops = {
    .proc_open    = continfo_open,
    .proc_read    = seq_read,
    .proc_lseek   = seq_lseek,
    .proc_release = single_release,
};

static int __init continfo_init(void)
{
    hash_init(cont_table);
    proc_create(PROC_NAME, 0444, NULL, &continfo_ops);
    printk(KERN_INFO "continfo_json modulo cargado\n");
    return 0;
}

static void __exit continfo_exit(void)
{
    remove_proc_entry(PROC_NAME, NULL);
    free_table();
    printk(KERN_INFO "continfo_json modulo desinstalado\n");
}

module_init(continfo_init);
module_exit(continfo_exit);
