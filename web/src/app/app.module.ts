import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { AppComponent } from './app.component';
import { HomeComponent } from './home/home.component';
import { AuthModule } from './auth/auth.module';
import { ComponentsModule } from './components/components.module';
import { MaterialModule } from './material/material.module';
import { MarkdownModule } from './markdown/markdown.module';

@NgModule({
  declarations: [

  ],
  imports: [
    BrowserModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    BrowserAnimationsModule,
    AuthModule,
    ComponentsModule,
    MaterialModule,
    MarkdownModule,
    AppComponent,
    HomeComponent
  ],
  providers: [],
  bootstrap: []
})
export class AppModule { }
