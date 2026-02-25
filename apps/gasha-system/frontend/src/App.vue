<template>
  <main class="app-shell">
    <div class="bg-orb bg-orb-a" aria-hidden="true" />
    <div class="bg-orb bg-orb-b" aria-hidden="true" />

    <header class="topbar glass">
      <div>
        <p class="eyebrow">Spec-driven UI</p>
        <h1>Gasha System</h1>
        <p class="subtle">登録からガシャ実行までを最短で確認できる検証フロント</p>
      </div>
      <nav class="tab-nav" aria-label="main navigation">
        <RouterLink to="/user" active-class="is-active">User</RouterLink>
        <RouterLink to="/admin" active-class="is-active">Admin</RouterLink>
      </nav>
    </header>

    <RouterView />
  </main>
</template>

<script setup lang="ts">
</script>
