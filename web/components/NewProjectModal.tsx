'use client';

import { useEffect, useState, useTransition } from "react";
import {
  FolderPlus,
  X,
  Sparkles,
  Folder,
  FolderOpen,
  Check,
  RefreshCw,
  HardDrive,
  ChevronRight,
  ChevronLeft,
  Search,
  Link2,
  CheckCircle2,
  AlertCircle
} from "lucide-react";

interface DriveFolderOption {
  id: string;
  name: string;
}

interface BreadcrumbItem {
  id: string;
  name: string;
}

interface NewProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated: (project: any) => void;
}

export default function NewProjectModal({ isOpen, onClose, onCreated }: NewProjectModalProps) {
  const [titre, setTitre] = useState("");
  const [mode, setMode] = useState<"browse" | "direct">("browse");

  // Navigation dans l'arborescence Drive
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
    { id: "root", name: "Mon Drive" }
  ]);
  const currentFolder = breadcrumbs[breadcrumbs.length - 1];

  // Dossier cible sélectionné pour enregistrer le projet
  const [selectedFolder, setSelectedFolder] = useState<{ id: string; name: string }>({
    id: "root",
    name: "AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO"
  });

  const [folders, setFolders] = useState<DriveFolderOption[]>([]);
  const [loadingFolders, setLoadingFolders] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [folderError, setFolderError] = useState<string | null>(null);

  // Saisie directe d'URL ou ID Drive
  const [directInput, setDirectInput] = useState("");
  const [validatingDirect, setValidatingDirect] = useState(false);
  const [directValidFolder, setDirectValidFolder] = useState<{ id: string; name: string } | null>(null);
  const [directError, setDirectError] = useState<string | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);
  const [companyProfile, setCompanyProfile] = useState<any>(null);
  const [initWorkspaceLoading, setInitWorkspaceLoading] = useState(false);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  // Charger le dossier 02_PROJETS_DAO par défaut et les dossiers Drive
  useEffect(() => {
    if (!isOpen) return;

    fetch(`${apiUrl}/api/company`, { credentials: "include" })
      .then((res) => res.json())
      .then((data) => {
        if (data) {
          setCompanyProfile(data);
          if (data.projects_folder_id) {
            setSelectedFolder({
              id: data.projects_folder_id,
              name: "AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO"
            });
          }
        }
      })
      .catch(() => {});

    loadDriveFolders(searchQuery ? "" : currentFolder.id, searchQuery);
  }, [isOpen, currentFolder.id, searchQuery]);

  const handleInitWorkspace = async () => {
    setInitWorkspaceLoading(true);
    try {
      const res = await fetch(`${apiUrl}/api/workspace/init`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      });
      const data = await res.json();
      if (res.ok && data.projects_folder_id) {
        setCompanyProfile(data);
        setSelectedFolder({
          id: data.projects_folder_id,
          name: "AGENTS_DAO_WORKSPACE / 02_PROJETS_DAO"
        });
        loadDriveFolders("root", "");
      }
    } catch (err) {
      console.error("Erreur init workspace:", err);
    } finally {
      setInitWorkspaceLoading(false);
    }
  };

  const loadDriveFolders = async (parentId = "root", query = "") => {
    setLoadingFolders(true);
    setFolderError(null);
    try {
      let url = `${apiUrl}/api/drive/folders`;
      const params = new URLSearchParams();
      if (query.trim()) {
        params.append("q", query.trim());
      } else if (parentId) {
        params.append("parent", parentId);
      }
      if (params.toString()) {
        url += `?${params.toString()}`;
      }

      const res = await fetch(url, { credentials: "include" });
      const data = await res.json();

      if (data.authenticated === false) {
        setAuthenticated(false);
        setFolders([]);
        return;
      }

      setAuthenticated(Boolean(data.authenticated));
      if (!res.ok) {
        setFolderError(data.error || "Impossible de charger vos dossiers Drive");
        setFolders([]);
      } else {
        setFolders(data.folders || []);
      }
    } catch {
      setFolderError("API injoignable");
      setFolders([]);
    } finally {
      setLoadingFolders(false);
    }
  };

  // Validation d'une URL ou d'un identifiant Drive collé
  const handleValidateDirectInput = async () => {
    const raw = directInput.trim();
    if (!raw) return;

    // Extraction de l'ID depuis une URL Google Drive (ex: /folders/123456... ou id=123456...) ou ID brut
    const match = raw.match(/[-\w]{25,}/);
    const folderId = match ? match[0] : raw;

    setValidatingDirect(true);
    setDirectError(null);
    setDirectValidFolder(null);

    try {
      const res = await fetch(`${apiUrl}/api/drive/folders?folder_id=${encodeURIComponent(folderId)}`, {
        credentials: "include"
      });
      const data = await res.json();

      if (res.ok && data.folder) {
        const found = { id: data.folder.id, name: data.folder.name };
        setDirectValidFolder(found);
        setSelectedFolder(found);
      } else {
        setDirectError(data.error || "Dossier introuvable ou inaccessible avec ce compte Google.");
      }
    } catch {
      setDirectError("Impossible de vérifier le dossier auprès de l'API.");
    } finally {
      setValidatingDirect(false);
    }
  };

  if (!isOpen) return null;

  // Entrer dans un sous-dossier
  const handleEnterFolder = (folder: DriveFolderOption) => {
    setSearchQuery("");
    setBreadcrumbs((prev) => [...prev, { id: folder.id, name: folder.name }]);
  };

  // Naviguer via le fil d'Ariane
  const handleJumpToBreadcrumb = (index: number) => {
    setSearchQuery("");
    setBreadcrumbs((prev) => prev.slice(0, index + 1));
  };

  // Remonter d'un niveau
  const handleGoUp = () => {
    if (breadcrumbs.length > 1) {
      setSearchQuery("");
      setBreadcrumbs((prev) => prev.slice(0, -1));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!titre.trim()) return;

    setLoading(true);
    setError(null);

    const driveParentId = selectedFolder.id === "root" ? "" : selectedFolder.id;

    try {
      const res = await fetch(`${apiUrl}/api/projects`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          titre: titre.trim(),
          drive_parent_id: driveParentId,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || "Erreur lors de la création du projet");
      }

      const project = await res.json();
      onCreated(project);
      onClose();
      setTitre("");
      setBreadcrumbs([{ id: "root", name: "Mon Drive" }]);
      setSelectedFolder({ id: "root", name: "Mon Drive (Racine)" });
    } catch (err: any) {
      setError(err.message || "Erreur de connexion au serveur");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white rounded-2xl shadow-2xl border border-slate-100 max-w-xl w-full overflow-hidden flex flex-col max-h-[90vh]">
        {/* En-tête du modal */}
        <div className="bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 text-white p-5 sm:p-6 relative shrink-0">
          <button
            onClick={onClose}
            className="absolute top-5 right-5 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors"
          >
            <X size={20} />
          </button>
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-blue-500/20 border border-blue-400/30 rounded-xl text-blue-400">
              <FolderPlus size={24} />
            </div>
            <div>
              <h3 className="text-lg font-bold">Nouveau Dossier d'Appel d'Offres</h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Créez l'espace projet et choisissez son emplacement sur Google Drive
              </p>
            </div>
          </div>
        </div>

        {/* Formulaire défilant */}
        <form onSubmit={handleSubmit} className="p-5 sm:p-6 space-y-5 overflow-y-auto flex-1">
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 text-red-700 text-xs rounded-xl flex items-center gap-2">
              <AlertCircle size={16} className="shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
              Intitulé ou Objet du Marché <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              required
              value={titre}
              onChange={(e) => setTitre(e.target.value)}
              placeholder="ex: Réhabilitation de la route nationale RN2 - Tronçon Thiès"
              className="w-full px-4 py-2.5 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm text-slate-900 placeholder-slate-400 shadow-xs"
            />
          </div>

          {/* Sélecteur d'emplacement Google Drive */}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
                Emplacement sur Google Drive
              </label>

              {/* Bascule mode arborescence / lien direct */}
              <div className="flex items-center gap-1 bg-slate-100 p-0.5 rounded-lg text-[11px] font-medium text-slate-600">
                <button
                  type="button"
                  onClick={() => setMode("browse")}
                  className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                    mode === "browse" ? "bg-white text-blue-700 font-semibold shadow-xs" : "hover:text-slate-900"
                  }`}
                >
                  Arborescence
                </button>
                <button
                  type="button"
                  onClick={() => setMode("direct")}
                  className={`px-2.5 py-1 rounded-md transition-all cursor-pointer ${
                    mode === "direct" ? "bg-white text-blue-700 font-semibold shadow-xs" : "hover:text-slate-900"
                  }`}
                >
                  Lien direct / ID
                </button>
              </div>
            </div>

            {mode === "browse" ? (
              <div className="border border-slate-200 rounded-xl bg-slate-50 p-3 space-y-2.5">
                {/* Barre d'outils : recherche & actualiser */}
                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <Search size={14} className="absolute left-3 top-2.5 text-slate-400" />
                    <input
                      type="text"
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      placeholder="Rechercher un dossier Drive..."
                      className="w-full pl-8 pr-3 py-1.5 bg-white border border-slate-200 rounded-lg text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    {searchQuery && (
                      <button
                        type="button"
                        onClick={() => setSearchQuery("")}
                        className="absolute right-2 top-2 text-slate-400 hover:text-slate-600 text-xs"
                      >
                        <X size={13} />
                      </button>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={() => loadDriveFolders(searchQuery ? "" : currentFolder.id, searchQuery)}
                    disabled={loadingFolders}
                    title="Actualiser la liste"
                    className="p-1.5 bg-white border border-slate-200 hover:bg-slate-100 rounded-lg text-slate-600 transition-colors cursor-pointer"
                  >
                    <RefreshCw size={14} className={loadingFolders ? "animate-spin text-blue-600" : ""} />
                  </button>
                </div>

                {/* Fil d'Ariane / Breadcrumbs (quand hors recherche) */}
                {!searchQuery && (
                  <div className="flex items-center gap-1.5 py-1 px-1 text-xs text-slate-600 overflow-x-auto">
                    {breadcrumbs.length > 1 && (
                      <button
                        type="button"
                        onClick={handleGoUp}
                        className="p-1 hover:bg-slate-200 rounded text-slate-700 flex items-center gap-0.5 cursor-pointer font-medium"
                        title="Dossier parent"
                      >
                        <ChevronLeft size={14} />
                      </button>
                    )}
                    {breadcrumbs.map((crumb, idx) => {
                      const isLast = idx === breadcrumbs.length - 1;
                      return (
                        <div key={crumb.id} className="flex items-center gap-1 shrink-0">
                          {idx > 0 && <ChevronRight size={12} className="text-slate-400" />}
                          <button
                            type="button"
                            onClick={() => handleJumpToBreadcrumb(idx)}
                            className={`px-1.5 py-0.5 rounded text-[11px] transition-colors cursor-pointer ${
                              isLast
                                ? "bg-blue-100/70 text-blue-800 font-bold"
                                : "hover:bg-slate-200 text-slate-600 hover:text-slate-900"
                            }`}
                          >
                            {crumb.name}
                          </button>
                        </div>
                      );
                    })}
                  </div>
                )}

                {/* Bouton pour sélectionner le dossier courant comme destination */}
                {!searchQuery && (
                  <div
                    onClick={() => setSelectedFolder({ id: currentFolder.id, name: currentFolder.name })}
                    className={`flex items-center justify-between p-2.5 rounded-lg border cursor-pointer transition-all ${
                      selectedFolder.id === currentFolder.id
                        ? "bg-blue-50 border-blue-400 text-blue-900 shadow-xs font-semibold"
                        : "bg-white border-slate-200 hover:bg-slate-100/70 text-slate-700"
                    }`}
                  >
                    <div className="flex items-center gap-2 text-xs truncate">
                      {currentFolder.id === "root" ? (
                        <HardDrive size={15} className="text-blue-600 shrink-0" />
                      ) : (
                        <FolderOpen size={15} className="text-blue-600 shrink-0" />
                      )}
                      <span className="truncate">
                        Choisir ce dossier : <strong>{currentFolder.name}</strong>
                      </span>
                    </div>
                    {selectedFolder.id === currentFolder.id && (
                      <Check size={16} className="text-blue-600 shrink-0" />
                    )}
                  </div>
                )}

                {/* Liste des sous-dossiers */}
                <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
                  {loadingFolders ? (
                    <div className="p-4 text-center text-xs text-slate-400 flex items-center justify-center gap-2">
                      <RefreshCw size={14} className="animate-spin text-blue-600" />
                      <span>Chargement des dossiers Drive...</span>
                    </div>
                  ) : authenticated === false ? (
                    <div className="p-4 text-center space-y-2">
                      <p className="text-xs text-slate-500">
                        Connectez votre compte Google pour parcourir votre Drive.
                      </p>
                      <a
                        href={`${apiUrl}/login`}
                        className="inline-flex items-center gap-2 px-3 py-1.5 bg-white border border-slate-300 hover:bg-slate-100 text-slate-800 text-xs font-semibold rounded-lg"
                      >
                        <HardDrive size={14} className="text-blue-600" />
                        Se connecter à Google Drive
                      </a>
                    </div>
                  ) : folderError ? (
                    <div className="p-3 text-xs text-red-600">{folderError}</div>
                  ) : folders.length === 0 ? (
                    <div className="p-4 text-xs text-slate-400 text-center">
                      {searchQuery
                        ? "Aucun dossier trouvé pour cette recherche."
                        : "Aucun sous-dossier à l'intérieur de ce répertoire."}
                    </div>
                  ) : (
                    folders.map((folder) => {
                      const isSelected = selectedFolder.id === folder.id;
                      return (
                        <div
                          key={folder.id}
                          className={`flex items-center justify-between p-2 rounded-lg border transition-all text-xs ${
                            isSelected
                              ? "bg-blue-50/80 border-blue-400 text-blue-900"
                              : "bg-white border-slate-200 hover:border-slate-300 text-slate-700"
                          }`}
                        >
                          {/* Zone de sélection du dossier */}
                          <div
                            onClick={() => setSelectedFolder({ id: folder.id, name: folder.name })}
                            onDoubleClick={() => handleEnterFolder(folder)}
                            className="flex items-center gap-2.5 truncate flex-1 cursor-pointer py-0.5"
                          >
                            <Folder size={15} className={isSelected ? "text-blue-600" : "text-indigo-500 shrink-0"} />
                            <span className="truncate font-medium">{folder.name}</span>
                          </div>

                          {/* Actions : sélection & explorer les sous-dossiers */}
                          <div className="flex items-center gap-1.5 shrink-0 ml-2">
                            {isSelected && <Check size={15} className="text-blue-600 mr-1" />}
                            <button
                              type="button"
                              onClick={() => handleEnterFolder(folder)}
                              className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 hover:text-slate-900 rounded text-[11px] font-semibold flex items-center gap-0.5 transition-colors cursor-pointer"
                              title="Ouvrir et explorer ce dossier"
                            >
                              <span>Ouvrir</span>
                              <ChevronRight size={12} />
                            </button>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            ) : (
              /* Mode saisie directe / URL Drive */
              <div className="border border-slate-200 rounded-xl bg-slate-50 p-4 space-y-3">
                <p className="text-xs text-slate-600">
                  Collez le lien complet ou l'identifiant du dossier Google Drive dans lequel vous souhaitez créer ce projet :
                </p>
                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <Link2 size={15} className="absolute left-3 top-2.5 text-slate-400" />
                    <input
                      type="text"
                      value={directInput}
                      onChange={(e) => setDirectInput(e.target.value)}
                      placeholder="https://drive.google.com/drive/folders/1LZQ6laP..."
                      className="w-full pl-9 pr-3 py-2 bg-white border border-slate-200 rounded-xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={handleValidateDirectInput}
                    disabled={validatingDirect || !directInput.trim()}
                    className="px-3.5 py-2 bg-slate-900 hover:bg-slate-800 disabled:opacity-50 text-white rounded-xl text-xs font-semibold flex items-center gap-1.5 cursor-pointer shrink-0"
                  >
                    {validatingDirect && <RefreshCw size={12} className="animate-spin" />}
                    <span>Valider</span>
                  </button>
                </div>

                {directError && (
                  <div className="p-2.5 bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg flex items-center gap-2">
                    <AlertCircle size={14} className="shrink-0" />
                    <span>{directError}</span>
                  </div>
                )}

                {directValidFolder && (
                  <div className="p-2.5 bg-emerald-50 border border-emerald-200 text-emerald-800 text-xs rounded-lg flex items-center gap-2">
                    <CheckCircle2 size={16} className="text-emerald-600 shrink-0" />
                    <span>
                      Dossier sélectionné : <strong>{directValidFolder.name}</strong>
                    </span>
                  </div>
                )}
              </div>
            )}

            {/* Bannière si AGENTS_DAO_WORKSPACE n'a pas encore été initialisé */}
            {companyProfile && !companyProfile.workspace_folder_id && (
              <div className="flex items-center justify-between p-3 bg-indigo-50 border border-indigo-200 rounded-xl text-xs gap-3">
                <div className="flex items-center gap-2">
                  <Sparkles size={16} className="text-indigo-600 shrink-0" />
                  <span className="text-indigo-950 font-medium">
                    L&apos;espace <strong>AGENTS_DAO_WORKSPACE</strong> n&apos;a pas encore été créé sur votre Drive.
                  </span>
                </div>
                <button
                  type="button"
                  onClick={handleInitWorkspace}
                  disabled={initWorkspaceLoading}
                  className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg font-bold text-[11px] flex items-center gap-1.5 cursor-pointer shrink-0 shadow-xs"
                >
                  {initWorkspaceLoading ? <RefreshCw size={11} className="animate-spin" /> : <Sparkles size={11} />}
                  <span>Créer le Workspace</span>
                </button>
              </div>
            )}

            {/* Récapitulatif de l'emplacement sélectionné */}
            <div className="p-3 bg-blue-50/70 border border-blue-200/80 rounded-xl flex items-center justify-between gap-3 text-xs">
              <div className="flex items-center gap-2 truncate">
                <CheckCircle2 size={16} className="text-blue-600 shrink-0" />
                <span className="text-slate-600 truncate">
                  Destination : <strong className="text-slate-900">{selectedFolder.name}</strong>
                  {selectedFolder.id !== "root" && (
                    <span className="font-mono text-[10px] text-slate-400 ml-1.5 truncate">
                      ({selectedFolder.id})
                    </span>
                  )}
                </span>
              </div>
              <span className="text-[11px] text-blue-700 font-semibold shrink-0">
                Prêt
              </span>
            </div>

            <p className="text-[11px] text-slate-500">
              L'arborescence normalisée (Offres Administrative, Technique, Financière) sera créée automatiquement dans ce dossier.
            </p>
          </div>

          {/* Boutons d'action */}
          <div className="pt-2 flex items-center justify-end gap-3 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2.5 text-sm font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-100 rounded-xl transition-colors cursor-pointer"
            >
              Annuler
            </button>
            <button
              type="submit"
              disabled={loading || !titre.trim()}
              className="flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-xl shadow-md shadow-blue-500/20 disabled:opacity-50 transition-all cursor-pointer"
            >
              {loading ? (
                <span>Création en cours...</span>
              ) : (
                <>
                  <Sparkles size={16} />
                  <span>Créer le Projet</span>
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}