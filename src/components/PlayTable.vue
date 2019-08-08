<template>
  <div>
    <div id="playerSelection">
      <span>No. of players: </span>
      <span class="select">
        <v-select :options="[3, 4, 5]" @input="playersSelected" placeholder="Choose..." />
      </span>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue, Prop } from 'vue-property-decorator';

import Player from './Player.vue';

function removeElementWithSelector(selector: string, timeout: number) {
  const elem: HTMLElement | null = document.querySelector(selector);
  if (elem === null) {
    return;
  }

  elem.style.opacity = '0';
  setTimeout(() => {
    if (elem.parentElement) {
      elem.parentElement.removeChild(elem);
    }
  }, timeout);
}

@Component({
  components: {
    Player,
  },
})
export default class PlayTable extends Vue {

  private playersSelected(value: number) {
    console.debug(`Players selected: ${value}`);
    removeElementWithSelector('#playerSelection', 1000);
  }
}
</script>

<style scoped>
#playerSelection {
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity ease 1s;
}

#playerSelection > span {
  margin: 5px;
}

#playerSelection > .select {
  width: 60vw;
}

@media screen and (min-width: 700px) {
  #playerSelection > .select {
    width: 27vw;
  }
}
</style>
