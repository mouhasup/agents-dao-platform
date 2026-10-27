// lib/api/dao.ts
// Toutes les fonctions liées à l'analyse de DAO

import { apiClient } from "./client";
import { DataAnalyseDAO } from "@/components/dashboard/types";

/**
 * Lance l'analyse d'un dossier Drive par les agents IA.
 * Correspond à la route GET /analyze du backend Go.
 */
export async function analyzeDAO(): Promise<DataAnalyseDAO> {
  return apiClient.get<DataAnalyseDAO>("/analyze");
}

/**
 * Lance l'analyse d'un fichier Drive spécifique par son ID.
 * Correspond à la route POST /analyze du backend Go (à implémenter).
 */
export async function analyzeDAOFile(fileId: string): Promise<DataAnalyseDAO> {
  return apiClient.post<DataAnalyseDAO>("/analyze", { file_id: fileId });
}
