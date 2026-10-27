package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"agents-dao-platform/model"
)

// DocxGenerator génère un document Word (.docx) professionnel conforme OpenXML (ECMA-376)
type DocxGenerator struct{}

func NewDocxGenerator() *DocxGenerator {
	return &DocxGenerator{}
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// GenerateMemoireDocx génère le fichier .docx en mémoire
func (g *DocxGenerator) GenerateMemoireDocx(project *model.Project, companyProfile *model.CompanyProfile, data *model.TechnicalOfferData) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 1. [Content_Types].xml
	if err := addZipFile(zipWriter, "[Content_Types].xml", getContentTypesXML()); err != nil {
		return nil, err
	}

	// 2. _rels/.rels
	if err := addZipFile(zipWriter, "_rels/.rels", getRootRelsXML()); err != nil {
		return nil, err
	}

	// 3. word/_rels/document.xml.rels
	if err := addZipFile(zipWriter, "word/_rels/document.xml.rels", getDocRelsXML()); err != nil {
		return nil, err
	}

	// 4. word/styles.xml
	if err := addZipFile(zipWriter, "word/styles.xml", getStylesXML()); err != nil {
		return nil, err
	}

	// 5. word/document.xml
	docXML := g.buildDocumentXML(project, companyProfile, data)
	if err := addZipFile(zipWriter, "word/document.xml", docXML); err != nil {
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addZipFile(zw *zip.Writer, name string, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(content))
	return err
}

func getContentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
}

func getRootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
}

func getDocRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
}

func getStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri" w:cs="Calibri"/>
        <w:sz w:val="22"/>
        <w:szCs w:val="22"/>
        <w:color w:val="1E293B"/>
        <w:lang w:val="fr-FR"/>
      </w:rPr>
    </w:rPrDefault>
    <w:pPrDefault>
      <w:pPr>
        <w:spacing w:line="276" w:lineRule="auto" w:after="120"/>
      </w:pPr>
    </w:pPrDefault>
  </w:docDefaults>

  <w:style w:type="paragraph" w:default="1" w:styleId="Normal">
    <w:name w:val="Normal"/>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Title">
    <w:name w:val="Title"/>
    <w:pPr>
      <w:spacing w:before="360" w:after="180"/>
      <w:jc w:val="center"/>
    </w:pPr>
    <w:rPr>
      <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
      <w:b/>
      <w:sz w:val="48"/>
      <w:szCs w:val="48"/>
      <w:color w:val="1E3A8A"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Subtitle">
    <w:name w:val="Subtitle"/>
    <w:pPr>
      <w:spacing w:before="120" w:after="240"/>
      <w:jc w:val="center"/>
    </w:pPr>
    <w:rPr>
      <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
      <w:sz w:val="26"/>
      <w:szCs w:val="26"/>
      <w:color w:val="475569"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/>
    <w:pPr>
      <w:spacing w:before="360" w:after="160"/>
    </w:pPr>
    <w:rPr>
      <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
      <w:b/>
      <w:sz w:val="32"/>
      <w:szCs w:val="32"/>
      <w:color w:val="1E3A8A"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Heading2">
    <w:name w:val="heading 2"/>
    <w:pPr>
      <w:spacing w:before="240" w:after="120"/>
    </w:pPr>
    <w:rPr>
      <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
      <w:b/>
      <w:sz w:val="26"/>
      <w:szCs w:val="26"/>
      <w:color w:val="2563EB"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Heading3">
    <w:name w:val="heading 3"/>
    <w:pPr>
      <w:spacing w:before="180" w:after="80"/>
    </w:pPr>
    <w:rPr>
      <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>
      <w:b/>
      <w:sz w:val="24"/>
      <w:szCs w:val="24"/>
      <w:color w:val="334155"/>
    </w:rPr>
  </w:style>
</w:styles>`
}

func (g *DocxGenerator) buildDocumentXML(project *model.Project, profile *model.CompanyProfile, data *model.TechnicalOfferData) string {
	companyName := "Notre Entreprise"
	companyAddress := ""
	companyNinea := ""
	if profile != nil {
		if profile.Nom != "" {
			companyName = profile.Nom
		}
		if profile.Adresse != "" {
			companyAddress = profile.Adresse
		}
		if profile.NINEA != "" {
			companyNinea = profile.NINEA
		}
	}

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>`)

	// ==================== PAGE DE GARDE ====================
	sb.WriteString(`
    <w:p><w:pPr><w:spacing w:before="400" w:after="200"/><w:jc w:val="center"/></w:pPr>
      <w:r><w:rPr><w:b/><w:sz w:val="24"/><w:color w:val="2563EB"/></w:rPr>
        <w:t xml:space="preserve">RÉPONSE À L'APPEL D'OFFRES PUBLIC</w:t>
      </w:r>
    </w:p>

    <w:p><w:pPr><w:pStyle w:val="Title"/></w:pPr>
      <w:r><w:t>MÉMOIRE TECHNIQUE</w:t></w:r>
    </w:p>
    <w:p><w:pPr><w:pStyle w:val="Subtitle"/></w:pPr>
      <w:r><w:t>OFFRE TECHNIQUE &amp; MÉTHODOLOGIQUE D'EXÉCUTION DES TRAVAUX</w:t></w:r>
    </w:p>

    <w:p><w:pPr><w:spacing w:before="300" w:after="300"/><w:jc w:val="center"/></w:pPr>
      <w:r><w:rPr><w:b/><w:sz w:val="26"/><w:color w:val="1E293B"/></w:rPr>
        <w:t xml:space="preserve">PROJET : ` + escapeXML(project.Titre) + `</w:t>
      </w:r>
    </w:p>
  `)

	// Encadré Informations Marché
	sb.WriteString(`
    <w:tbl>
      <w:tblPr>
        <w:tblW w:w="5000" w:type="pct"/>
        <w:tblBorders>
          <w:top w:val="single" w:sz="8" w:space="0" w:color="2563EB"/>
          <w:left w:val="single" w:sz="8" w:space="0" w:color="2563EB"/>
          <w:bottom w:val="single" w:sz="8" w:space="0" w:color="2563EB"/>
          <w:right w:val="single" w:sz="8" w:space="0" w:color="2563EB"/>
          <w:insideH w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
          <w:insideV w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
        </w:tblBorders>
      </w:tblPr>
      <w:tr>
        <w:tc><w:tcPr><w:tcW w:w="1800" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="F8FAFC"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="475569"/></w:rPr><w:t>Autorité Contractante :</w:t></w:r></w:p>
        </w:tc>
        <w:tc><w:tcPr><w:tcW w:w="3200" w:type="pct"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>` + escapeXML(fallback(project.Acheteur, "Non spécifié")) + `</w:t></w:r></w:p>
        </w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:tcPr><w:tcW w:w="1800" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="F8FAFC"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="475569"/></w:rPr><w:t>Référence / N° DAO :</w:t></w:r></w:p>
        </w:tc>
        <w:tc><w:tcPr><w:tcW w:w="3200" w:type="pct"/></w:tcPr>
          <w:p><w:r><w:rPr><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>` + escapeXML(fallback(project.NumeroDAO, "Dossier #"+project.ID)) + `</w:t></w:r></w:p>
        </w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:tcPr><w:tcW w:w="1800" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="F8FAFC"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="475569"/></w:rPr><w:t>Entreprise Soumissionnaire :</w:t></w:r></w:p>
        </w:tc>
        <w:tc><w:tcPr><w:tcW w:w="3200" w:type="pct"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="1E3A8A"/></w:rPr><w:t>` + escapeXML(companyName) + `</w:t></w:r>`)

	if companyNinea != "" {
		sb.WriteString(`<w:r><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t xml:space="preserve"> (NINEA : ` + escapeXML(companyNinea) + `)</w:t></w:r>`)
	}
	if companyAddress != "" {
		sb.WriteString(`<w:r><w:rPr><w:sz w:val="18"/><w:color w:val="64748B"/></w:rPr><w:t xml:space="preserve"> - ` + escapeXML(companyAddress) + `</w:t></w:r>`)
	}

	sb.WriteString(`</w:p></w:tc>
      </w:tr>
      <w:tr>
        <w:tc><w:tcPr><w:tcW w:w="1800" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="F8FAFC"/></w:tcPr>
          <w:p><w:r><w:rPr><w:b/><w:sz w:val="20"/><w:color w:val="475569"/></w:rPr><w:t>Date d'Édition :</w:t></w:r></w:p>
        </w:tc>
        <w:tc><w:tcPr><w:tcW w:w="3200" w:type="pct"/></w:tcPr>
          <w:p><w:r><w:rPr><w:sz w:val="20"/><w:color w:val="1E293B"/></w:rPr><w:t>` + time.Now().Format("02 Janvier 2006") + `</w:t></w:r></w:p>
        </w:tc>
      </w:tr>
    </w:tbl>
  `)

	// Saut de page après page de garde
	sb.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)

	// ==================== SOMMAIRE ====================
	sb.WriteString(`
    <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
      <w:r><w:t>SOMMAIRE GÉNÉRAL DU MÉMOIRE TECHNIQUE</w:t></w:r>
    </w:p>
    <w:p><w:pPr><w:spacing w:after="200"/></w:pPr>
      <w:r><w:rPr><w:i/><w:color w:val="64748B"/></w:rPr>
        <w:t>Le présent mémoire technique est structuré selon les exigences de la réglementation des marchés publics et les spécifications techniques du DAO.</w:t>
      </w:r>
    </w:p>
  `)

	for i, sec := range data.Sections {
		sb.WriteString(fmt.Sprintf(`
      <w:p><w:pPr><w:spacing w:before="60" w:after="60"/></w:pPr>
        <w:r><w:rPr><w:b/><w:color w:val="1E3A8A"/></w:rPr><w:t xml:space="preserve">Chapitre %d : </w:t></w:r>
        <w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r>
      </w:p>
    `, i+1, escapeXML(sec.Titre)))
	}

	if len(data.Personnel) > 0 {
		sb.WriteString(`
      <w:p><w:pPr><w:spacing w:before="80" w:after="60"/></w:pPr>
        <w:r><w:rPr><w:b/><w:color w:val="2563EB"/></w:rPr><w:t xml:space="preserve">Annexe A : </w:t></w:r>
        <w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>Personnel Clé &amp; Encadrement Mobilisé</w:t></w:r>
      </w:p>
    `)
	}

	if len(data.Materiel) > 0 {
		sb.WriteString(`
      <w:p><w:pPr><w:spacing w:before="60" w:after="60"/></w:pPr>
        <w:r><w:rPr><w:b/><w:color w:val="2563EB"/></w:rPr><w:t xml:space="preserve">Annexe B : </w:t></w:r>
        <w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>Moyens Matériels &amp; Engins de Chantier Alloués</w:t></w:r>
      </w:p>
    `)
	}

	// Saut de page après sommaire
	sb.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)

	// ==================== SECTIONS DU MÉMOIRE ====================
	for i, sec := range data.Sections {
		sb.WriteString(fmt.Sprintf(`
      <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
        <w:r><w:t>Chapitre %d. %s</w:t></w:r>
      </w:p>
    `, i+1, escapeXML(sec.Titre)))

		if strings.TrimSpace(sec.Contenu) != "" {
			g.renderMarkdownParagraphs(&sb, sec.Contenu)
		} else {
			sb.WriteString(`
        <w:p>
          <w:r><w:rPr><w:i/><w:color w:val="94A3B8"/></w:rPr>
            <w:t>[Section en attente de rédaction]</w:t>
          </w:r>
        </w:p>
      `)
		}

		sb.WriteString(`<w:p><w:pPr><w:spacing w:before="240" w:after="240"/></w:pPr></w:p>`)
	}

	// ==================== ANNEXE 1 : PERSONNEL ====================
	if len(data.Personnel) > 0 {
		sb.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		sb.WriteString(`
      <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
        <w:r><w:t>Annexe A : Personnel Clé &amp; Encadrement Technique Mobilisé</w:t></w:r>
      </w:p>
      <w:p><w:pPr><w:spacing w:after="160"/></w:pPr>
        <w:r><w:rPr><w:i/><w:color w:val="64748B"/></w:rPr>
          <w:t>Le tableau ci-dessous synthétise l'équipe d'experts et d'encadrement affectée à la réalisation du chantier, avec leurs qualifications et leur expérience professionnelle :</w:t>
        </w:r>
      </w:p>

      <w:tbl>
        <w:tblPr>
          <w:tblW w:w="5000" w:type="pct"/>
          <w:tblBorders>
            <w:top w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:left w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:bottom w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:right w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:insideH w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
            <w:insideV w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
          </w:tblBorders>
        </w:tblPr>
        <w:tr>
          <w:tc><w:tcPr><w:tcW w:w="1400" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Rôle Requis (DAO)</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1300" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Nom de l'Expert</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="700" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Expérience</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1600" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Diplômes &amp; Spécialités</w:t></w:r></w:p>
          </w:tc>
        </w:tr>
    `)

		for _, p := range data.Personnel {
			sb.WriteString(fmt.Sprintf(`
        <w:tr>
          <w:tc><w:tcPr><w:tcW w:w="1400" w:type="pct"/></w:tcPr>
            <w:p><w:r><w:rPr><w:b/><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1300" w:type="pct"/></w:tcPr>
            <w:p><w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="700" w:type="pct"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>%d ans</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1600" w:type="pct"/></w:tcPr>
            <w:p><w:r><w:rPr><w:color w:val="475569"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
        </w:tr>
      `, escapeXML(p.RoleRequis), escapeXML(p.Nom), p.ExperienceAnnees, escapeXML(p.Qualifications)))
		}
		sb.WriteString(`</w:tbl>`)
	}

	// ==================== ANNEXE 2 : MATÉRIEL ====================
	if len(data.Materiel) > 0 {
		sb.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		sb.WriteString(`
      <w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr>
        <w:r><w:t>Annexe B : Moyens Matériels &amp; Engins Mobilisés</w:t></w:r>
      </w:p>
      <w:p><w:pPr><w:spacing w:after="160"/></w:pPr>
        <w:r><w:rPr><w:i/><w:color w:val="64748B"/></w:rPr>
          <w:t>Le tableau ci-dessous présente les équipements et engins de chantier affectés pour garantir le respect du planning et des cadences d'exécution :</w:t>
        </w:r>
      </w:p>

      <w:tbl>
        <w:tblPr>
          <w:tblW w:w="5000" w:type="pct"/>
          <w:tblBorders>
            <w:top w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:left w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:bottom w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:right w:val="single" w:sz="6" w:space="0" w:color="CBD5E1"/>
            <w:insideH w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
            <w:insideV w:val="single" w:sz="4" w:space="0" w:color="E2E8F0"/>
          </w:tblBorders>
        </w:tblPr>
        <w:tr>
          <w:tc><w:tcPr><w:tcW w:w="1500" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Matériel Exigé (DAO)</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1500" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Désignation &amp; Modèle</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1000" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Immatriculation / N°</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1000" w:type="pct"/><w:shd w:val="clear" w:color="auto" w:fill="1E3A8A"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:color w:val="FFFFFF"/></w:rPr><w:t>Statut Juridique</w:t></w:r></w:p>
          </w:tc>
        </w:tr>
    `)

		for _, m := range data.Materiel {
			sb.WriteString(fmt.Sprintf(`
        <w:tr>
          <w:tc><w:tcPr><w:tcW w:w="1500" w:type="pct"/></w:tcPr>
            <w:p><w:r><w:rPr><w:b/><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1500" w:type="pct"/></w:tcPr>
            <w:p><w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1000" w:type="pct"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:color w:val="1E293B"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
          <w:tc><w:tcPr><w:tcW w:w="1000" w:type="pct"/></w:tcPr>
            <w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:color w:val="475569"/></w:rPr><w:t>%s</w:t></w:r></w:p>
          </w:tc>
        </w:tr>
      `, escapeXML(m.TypeRequis), escapeXML(m.Designation), escapeXML(fallback(m.Immatriculation, "-")), escapeXML(m.Statut)))
		}
		sb.WriteString(`</w:tbl>`)
	}

	sb.WriteString(`
  </w:body>
</w:document>`)

	return sb.String()
}

func (g *DocxGenerator) renderMarkdownParagraphs(sb *strings.Builder, text string) {
	lines := strings.Split(text, "\n")
	boldRegex := regexp.MustCompile(`\*\*(.*?)\*\*`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Titre niveau 2 Markdown
		if strings.HasPrefix(trimmed, "## ") {
			titleText := strings.TrimPrefix(trimmed, "## ")
			sb.WriteString(fmt.Sprintf(`
        <w:p><w:pPr><w:pStyle w:val="Heading2"/></w:pPr>
          <w:r><w:t>%s</w:t></w:r>
        </w:p>
      `, escapeXML(titleText)))
			continue
		}

		// Titre niveau 3 Markdown
		if strings.HasPrefix(trimmed, "### ") {
			titleText := strings.TrimPrefix(trimmed, "### ")
			sb.WriteString(fmt.Sprintf(`
        <w:p><w:pPr><w:pStyle w:val="Heading3"/></w:pPr>
          <w:r><w:t>%s</w:t></w:r>
        </w:p>
      `, escapeXML(titleText)))
			continue
		}

		// Puce Markdown (- ou * ou •)
		isBullet := strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "• ")
		if isBullet {
			content := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* "), "• ")
			sb.WriteString(`
        <w:p><w:pPr><w:spacing w:before="40" w:after="40"/><w:ind w:left="400" w:hanging="200"/></w:pPr>
          <w:r><w:rPr><w:color w:val="2563EB"/><w:b/></w:rPr><w:t xml:space="preserve">▪ </w:t></w:r>
      `)
			g.renderInlineRuns(sb, content, boldRegex)
			sb.WriteString(`</w:p>`)
			continue
		}

		// Paragraphe standard
		sb.WriteString(`<w:p><w:pPr><w:spacing w:before="60" w:after="100"/></w:pPr>`)
		g.renderInlineRuns(sb, trimmed, boldRegex)
		sb.WriteString(`</w:p>`)
	}
}

func (g *DocxGenerator) renderInlineRuns(sb *strings.Builder, text string, boldRegex *regexp.Regexp) {
	indices := boldRegex.FindAllStringSubmatchIndex(text, -1)
	if len(indices) == 0 {
		sb.WriteString(fmt.Sprintf(`<w:r><w:t xml:space="preserve">%s</w:t></w:r>`, escapeXML(text)))
		return
	}

	lastIdx := 0
	for _, idx := range indices {
		// Texte normal avant le gras
		if idx[0] > lastIdx {
			normalText := text[lastIdx:idx[0]]
			sb.WriteString(fmt.Sprintf(`<w:r><w:t xml:space="preserve">%s</w:t></w:r>`, escapeXML(normalText)))
		}

		// Texte en gras
		boldText := text[idx[2]:idx[3]]
		sb.WriteString(fmt.Sprintf(`<w:r><w:rPr><w:b/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`, escapeXML(boldText)))

		lastIdx = idx[1]
	}

	if lastIdx < len(text) {
		remaining := text[lastIdx:]
		sb.WriteString(fmt.Sprintf(`<w:r><w:t xml:space="preserve">%s</w:t></w:r>`, escapeXML(remaining)))
	}
}

func fallback(val, def string) string {
	if strings.TrimSpace(val) == "" {
		return def
	}
	return val
}
