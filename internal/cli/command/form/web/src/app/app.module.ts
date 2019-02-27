import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { MatCardModule } from '@angular/material/card';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { HttpClientModule } from '@angular/common/http';
import { UniformModule } from '@banzaicloud/uniform';

import { AppComponent } from './app.component';

@NgModule({
  declarations: [AppComponent],
  imports: [BrowserModule, HttpClientModule, UniformModule, MatCardModule, MatProgressBarModule, MatSnackBarModule],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
