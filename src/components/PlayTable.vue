<template>
  <div>
    <div id="playerSelection">
      <span>No. of players: </span>
      <span class="select">
        <v-select :options="[3, 4, 5]" @input="playersSelected" placeholder="Choose..." />
      </span>
    </div>
    <div id="tableArea" v-bind:style="styles">
      <Player v-for="p in players"
              v-bind:key="p.idx"
              v-bind:idx="p.idx"
              v-bind:numPlayers="p.total"
              v-bind:canMove="p.moveable"
              v-bind:tableSize="tableSize" />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

import Player from './Player.vue';

/**
 * Removes the element matching the given selector by zero'ing its
 * opacity and removing itself from its parent.
 *
 * @param selector The query selector used for finding the element.
 * @param timeout Timeout for the element to smoothly disappear
 * from the screen (should match the CSS transition).
 */
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

/**
 * Timeout (ms) for the selection dropdown to disappear and player
 * placement to begin.
 */
const PLAYER_PLACING_TIMEOUT = 1000;

@Component({
  components: {
    Player,
  },
})
export default class PlayTable extends Vue {

  private players: object[] = [];

  /**
   * @returns The size of the table in viewport width unit (vw).
   */
  private get tableSize(): number {
    let width = 65;
    if (screen.width > 700) {
      width = 30;
    }

    return width;
  }

  /**
   * @returns The styles associated with this component.
   */
  private get styles(): object {
    const size = this.tableSize;
    return {
      width: `${size}vw`,
      height: `${size}vw`,
    };
  }

  /**
   * Triggered once the user has selected the no. of players.
   * This adds child (`Player`) components.
   *
   * @param total No. of players in this table.
   */
  private playersSelected(total: number) {
    console.debug(`Players selected: ${total}`);
    removeElementWithSelector('#playerSelection', PLAYER_PLACING_TIMEOUT);

    setTimeout(() => {
      for (const idx of Array(total).keys()) {
        this.players.push({ idx, total, moveable: false });
        setTimeout(() => {
          this.players[idx]["moveable"] = true;
        }, (idx + 1) * 250);
      }
    }, PLAYER_PLACING_TIMEOUT);
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

#tableArea {
  margin: 10% 12.5vw 0 12.5vw;
  width: 65vw;
  height: 65vw;
}

@media screen and (min-width: 700px) {
  #playerSelection > .select {
    width: 27vw;
  }
}
</style>
