import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { NgIcon, provideIcons } from '@ng-icons/core';
import { lucideFileText, lucidePackage } from '@ng-icons/lucide';

import { ZardSidebarImports } from '@/shared/components/sidebar';
import { ZardSonnerComponent } from '@/shared/components/sonner';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink, RouterLinkActive, NgIcon, ZardSonnerComponent, ...ZardSidebarImports],
  providers: [provideIcons({ lucidePackage, lucideFileText })],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {}
