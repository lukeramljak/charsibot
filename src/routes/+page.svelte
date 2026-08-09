<script>
  import { resolve } from '$app/paths';
</script>

<div class="flex min-h-screen items-center justify-center bg-white">
  <div class="text-center">
    <div class="flex flex-col gap-4">
      <a
        href={resolve('/blind-box')}
        class="rounded-lg bg-black px-8 py-4 font-semibold text-white transition-colors duration-200 hover:bg-black/80"
      >
        Blind Box
      </a>
      <a
        href={resolve('/admin')}
        class="rounded-lg bg-slate-700 px-8 py-4 font-semibold text-white transition-colors duration-200 hover:bg-slate-600"
      >
        Admin
      </a>
    </div>
  </div>
</div>
