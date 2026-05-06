<script setup lang="ts">
import '@/assets/styles/menu/menu-item.css'

const BRACKET_LEFT = '\u003E'
const BRACKET_RIGHT = '\u003C'

const props = defineProps<{
  label: string
  selected: boolean
  kind: 'action' | 'cycler'
}>()

const emit = defineEmits<{
  (e: 'activate'): void
  (e: 'select'): void
}>()

function onClick() {
  emit('select')
  if (props.kind === 'action') {
    emit('activate')
  }
}
</script>

<template>
  <li
    class="menu-item"
    :class="{
      'menu-item--selected': selected,
      'menu-item--action': kind === 'action',
      'menu-item--cycler': kind === 'cycler',
    }"
    role="menuitem"
    :aria-current="selected ? 'true' : undefined"
    @click="onClick"
  >
    <span class="menu-item__bracket menu-item__bracket--left" aria-hidden="true">
      {{ selected ? BRACKET_LEFT : '' }}
    </span>
    <span class="menu-item__label">
      <span class="menu-item__label-text">{{ label }}</span>
      <slot />
    </span>
    <span class="menu-item__bracket menu-item__bracket--right" aria-hidden="true">
      {{ selected ? BRACKET_RIGHT : '' }}
    </span>
  </li>
</template>
