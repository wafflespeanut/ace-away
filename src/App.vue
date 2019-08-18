<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawerOpen" app>
      <!--  -->
    </v-navigation-drawer>

    <v-app-bar app>
      <v-app-bar-nav-icon @click.stop="drawerOpen = !drawerOpen"></v-app-bar-nav-icon>
      <v-toolbar-title>Ace Away!</v-toolbar-title>
    </v-app-bar>

    <v-content>
      <v-container class="fill-height" fluid>
        <!--  -->
      </v-container>
      <CreateRoom :players="allowedPlayers" :showDialog="isNewSession" />
    </v-content>
    <v-footer app>
      <span class="white--text">&copy; 2019 @wafflespeanut</span>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';
import CreateRoom from './dialog/CreateRoom.vue';
import { ClientMessage, RoomCreationRequest } from './persistence/model';

const ALLOWED_PLAYERS: number[] = [3, 4, 5, 6];

@Component({
  components: {
    CreateRoom,
  },
})
export default class App extends Vue {
  /**
   * @returns whether the player is creating a new room.
   */
  private get isNewSession(): boolean {
    return this.roomCreated || window.location.pathname.indexOf('join') === -1;
  }

  private drawerOpen: boolean = false;

  private roomCreated: boolean = false;

  private allowedPlayers: number[] = ALLOWED_PLAYERS;
}
export { ALLOWED_PLAYERS };
</script>
