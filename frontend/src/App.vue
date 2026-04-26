<script setup lang="ts">
import { NConfigProvider, NMessageProvider, NDialogProvider, NLoadingBarProvider, NGlobalStyle, darkTheme } from 'naive-ui'
import { RouterView } from 'vue-router'
import { removeFileDropListener } from './utils/fileDrop'
import { onUnmounted } from 'vue'

onUnmounted(() => {
  removeFileDropListener()
})
</script>

<template>
  <NConfigProvider abstract :theme="darkTheme">
    <NGlobalStyle />
    <NMessageProvider>
      <NDialogProvider>
        <NLoadingBarProvider>
          <RouterView />
        </NLoadingBarProvider>
      </NDialogProvider>
    </NMessageProvider>
  </NConfigProvider>
</template>

<style scoped>
</style>
