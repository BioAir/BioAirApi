import { Injectable } from '@angular/core';
import { HttpClient} from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class BioairService {

  private baseURL = 'http://localhost:9100';


  
  constructor(private http: HttpClient) { }

  getCarterasByUser(){
    return this.http.get(`${this.baseURL}/aire_data`);
  }
}
