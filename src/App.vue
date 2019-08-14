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
      <v-dialog v-model="dialogShow" persistent>
        <v-card>
          <v-card-title class="headline">{{ dialogHeading }}</v-card-title>
          <v-card-text>{{ dialogMessage }}</v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn v-for="(button, i) in dialogButtons" v-bind:key="i"
                   color="red darken-1" text
                   @click="button.handler">{{ button.content }}</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-content>
    <v-footer app>
      <span class="white--text">&copy; 2019 @wafflespeanut</span>
    </v-footer>
  </v-app>
</template>

<script lang="ts">
import Vue from 'vue';
import Component from 'vue-class-component';

interface DialogButton {
  handler: () => void;
  content: string;
}

@Component({
  components: {},
})
export default class App extends Vue {

  private source: string = '';

  private dialogShow: boolean = false;

  private dialogHeading: string = '';

  private dialogMessage: string = '';

  private drawerOpen: boolean = false;

  private dialogButtons: DialogButton[] = [];

  private mounted() {
    this.prepareForRoomCreation();
  }

  /**
   * Sets the dialogs, messages and buttons for creating a new room.
   */
  private prepareForRoomCreation() {
    this.dialogHeading = 'Create Room';
    this.dialogMessage = 'Create a new room for playing.';
    this.dialogButtons.push({
      handler: () => {
        this.dialogShow = false;
      },
      content: 'Create',
    });
    this.dialogShow = true;
  }
}
</script>
