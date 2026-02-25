<template>
  <main class="container">
    <header class="topbar">
      <h1>Gasha System</h1>
      <nav>
        <RouterLink to="/user">User</RouterLink>
        <RouterLink to="/admin">Admin</RouterLink>
      </nav>
    </header>
    <RouterView />
  </main>
</template>

<script setup lang="ts">
</script>
