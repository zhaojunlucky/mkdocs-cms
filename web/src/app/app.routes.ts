import { Routes } from '@angular/router';
import { LoginComponent } from './auth/login/login.component';
import { HomeComponent } from './home/home.component';
import { RepositoryDetailComponent } from './repositories/repository-detail/repository-detail.component';
import { RepositoryImportComponent } from './repositories/repository-import/repository-import.component';
import { EditRepositoryComponent } from './repositories/edit-repository/edit-repository.component';
import { CollectionComponent } from './repositories/collection/collection.component';
import { EditFileComponent } from './repositories/edit-file/edit-file.component';
import { CreateFileComponent } from './repositories/create-file/create-file.component';
import { authGuard } from './auth/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/home', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'home', component: HomeComponent, canActivate: [authGuard] },
  { path: 'repositories/import', component: RepositoryImportComponent, canActivate: [authGuard] },
  { path: 'repositories/:id', component: RepositoryDetailComponent, canActivate: [authGuard] ,
    children: [
      { path: 'collection/:collectionName', component: CollectionComponent, canActivate: [authGuard] },
      { path: 'collection/:collectionName/**', component: CollectionComponent, canActivate: [authGuard] },
      { path: 'collection/:collectionName/edit/**', component: EditFileComponent, canActivate: [authGuard] },
      { path: 'collection/:collectionName/create/**', component: CreateFileComponent, canActivate: [authGuard] },
    ]
  },
  { path: 'repositories/:id/edit', component: EditRepositoryComponent, canActivate: [authGuard] },

  { path: '**', redirectTo: '/home' }
];
