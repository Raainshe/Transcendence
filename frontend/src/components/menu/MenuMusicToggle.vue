<script setup lang="ts">
import AudioMuteToggle from '@/components/menu/AudioMuteToggle.vue'
</script>

<template>
  <AudioMuteToggle target="music" class="menu-music-toggle-slot" />
</template>

<style scoped>
.menu-music-toggle-slot {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
}
</style>
