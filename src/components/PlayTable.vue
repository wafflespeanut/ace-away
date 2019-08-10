<template>
  <div>
    <div id="banner">{{ bannerMsg }}</div>
    <div id="message">{{ msg }}</div>
    <div id="playerSelection">
      <span>No. of players: </span>
      <span class="select">
        <v-select :options="[3, 4, 5, 6]" @input="playersSelected" placeholder="Choose..." />
      </span>
    </div>
    <div id="cardSelection" v-if="hasPlayers">
      <div>
        <span class="suite" v-for="(suite, i) in suites" v-bind:key="i">
          {{ suite }}
        </span>
      </div>
      <div>
        <span class="label" v-for="(label, i) in labels" v-bind:key="i">
          {{ label }}
        </span>
      </div>
    </div>
    <div id="tableArea" v-bind:style="styles">
      <Player v-for="(p, idx) in players"
              v-bind:key="idx"
              v-bind:idx="idx"
              v-bind:numPlayers="p.total"
              v-bind:canMove="p.moveable"
              v-bind:tableProps="tableProps" />
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
/**
 * Timeout (ms) for messages fading in/out.
 */
const MSG_TIMEOUT = 1000;

interface TableProperties {
  width: number;
  height: number;
  offsetX: number;
  offsetY: number;
}

@Component({
  components: {
    Player,
  },
})
export default class PlayTable extends Vue {

  private players: object[] = [];

  private msg: string = '';

  private bannerMsg: string = '';

  private readonly suites: string[] = ['♣', '♦', '♥', '♠'];

  private readonly labels: string[] = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A'];

  /**
   * @returns Some properties of the table in viewport width units (vw).
   */
  private get tableProps(): TableProperties {
    const offsetX = 0;
    const width = 65;
    let height = 40;
    let offsetY = 18;
    if (screen.width > 700) {
      height = 55;
      offsetY = 20;
    }

    return {
      width,
      height,
      offsetX,
      offsetY,
    };
  }

  private get hasPlayers(): boolean {
    return this.players.length > 0;
  }

  /**
   * @returns The styles associated with this component.
   */
  private get styles(): object {
    return {};
  }

  public displayMessage(banner: string, msg: string) {
    let msgEl = this.$el.querySelector('#message');
    let banEl = this.$el.querySelector('#banner');
    msgEl.style.opacity = '0';
    banEl.style.opacity = '0';

    setTimeout(() => {
      this.msg = msg;
      this.bannerMsg = banner;
      msgEl.style.opacity = '1';
      banEl.style.opacity = '1';
    }, MSG_TIMEOUT);
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
    this.displayMessage('Add your cards.', 'Pick a suite and a label to add your card.');

    setTimeout(() => {
      for (const idx of Array(total).keys()) {
        this.players.push({ total, moveable: false });
        setTimeout(() => {
          this.players[idx].moveable = true;
        }, (idx + 1) * 250);
      }
    }, PLAYER_PLACING_TIMEOUT);
  }
}

export { TableProperties };
</script>

<style scoped>
#banner {
  font-size: 3.5vh;
}

#message, #banner {
  opacity: 0;
  transition: opacity ease 1s;
}

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

#cardSelection .label {
  margin: 5px;
  font-size: 4vh;
}

#cardSelection .suite {
  font-size: 9vh;
}

@media screen and (min-width: 700px) {
  #banner {
    font-size: 2.75vh;
  }

  #playerSelection > .select {
    width: 27vw;
  }
}
</style>
