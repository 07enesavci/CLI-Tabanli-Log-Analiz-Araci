package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"log-analyzer/backend/internal/analyzer"
	"log-analyzer/backend/internal/rules"
	"log-analyzer/backend/internal/tailer"
)

func main() {
	configPath := "config/rules.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	ruleManager, err := rules.NewManager(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	analyzer := analyzer.NewAnalyzer(ruleManager)
	tailer := tailer.NewTailer(ruleManager)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print("Seçiminiz: ")
		
		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			analyzeFiles(analyzer, ruleManager)
		case "2":
			startTailing(tailer, ruleManager)
		case "3":
			viewRules(ruleManager)
		case "4":
			viewLogFiles(ruleManager)
		case "5":
			fmt.Println("Çıkılıyor...")
			tailer.Stop()
			return
		default:
			fmt.Println("Geçersiz seçim! Lütfen 1-5 arası bir sayı girin.")
		}
	}
}

func printMenu() {
	fmt.Println("\n=== Log Analiz ve Uyarı Aracı ===")
	fmt.Println("1. Dosya Bazlı Analiz")
	fmt.Println("2. Gerçek Zamanlı İzleme (Tailing)")
	fmt.Println("3. Kuralları Görüntüle")
	fmt.Println("4. Log Dosyalarını Görüntüle")
	fmt.Println("5. Çıkış")
	fmt.Println()
}

func analyzeFiles(analyzer *analyzer.Analyzer, ruleManager *rules.Manager) {
	fmt.Println("\n=== Dosya Bazlı Analiz ===")
	
	logFiles := ruleManager.GetEnabledLogFiles()
	if len(logFiles) == 0 {
		fmt.Println("Etkin log dosyası bulunamadı!")
		return
	}

	fmt.Println("\nEtkin log dosyaları:")
	for i, file := range logFiles {
		fmt.Printf("%d. %s (%s)\n", i+1, file.Path, file.Type)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\nAnaliz edilecek dosyaları seçin (virgülle ayırın, 'all' için tümünü seçin): ")
	scanner.Scan()
	selection := strings.TrimSpace(scanner.Text())

	var filesToAnalyze []string
	if strings.ToLower(selection) == "all" {
		for _, file := range logFiles {
			filesToAnalyze = append(filesToAnalyze, file.Path)
		}
	} else {
		indices := strings.Split(selection, ",")
		for _, idxStr := range indices {
			var idx int
			if _, err := fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx); err == nil && idx > 0 && idx <= len(logFiles) {
				filesToAnalyze = append(filesToAnalyze, logFiles[idx-1].Path)
			}
		}
	}

	if len(filesToAnalyze) == 0 {
		fmt.Println("Geçerli dosya seçilmedi!")
		return
	}

	fmt.Println("\nAnaliz başlatılıyor...")
	entries, err := analyzer.AnalyzeMultipleFiles(filesToAnalyze)
	if err != nil {
		fmt.Printf("Analiz hatası: %v\n", err)
		return
	}

	fmt.Printf("\nToplam %d uyarı bulundu.\n", len(entries))
	severityCount := make(map[string]int)
	for _, entry := range entries {
		severityCount[entry.Severity]++
	}

	fmt.Println("\nÖzet:")
	for severity, count := range severityCount {
		fmt.Printf("  %s: %d\n", severity, count)
	}
	fmt.Print("\nCSV olarak kaydetmek ister misiniz? (e/h): ")
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "e" {
		fmt.Print("Dosya adı (örn: report.csv): ")
		scanner.Scan()
		outputPath := strings.TrimSpace(scanner.Text())
		if outputPath == "" {
			outputPath = "report.csv"
		}

		if err := analyzer.ExportToCSV(entries, outputPath); err != nil {
			fmt.Printf("CSV kaydetme hatası: %v\n", err)
		} else {
			fmt.Printf("Rapor %s dosyasına kaydedildi.\n", outputPath)
		}
	}
	fmt.Println("\nSon 10 uyarı:")
	start := 0
	if len(entries) > 10 {
		start = len(entries) - 10
	}
	for i := start; i < len(entries); i++ {
		entry := entries[i]
		fmt.Printf("\n[%s] %s - %s\n", entry.Severity, entry.Timestamp, strings.Join(entry.MatchedRules, ", "))
		fmt.Printf("  Dosya: %s\n", entry.Source)
		fmt.Printf("  Satır: %s\n", truncate(entry.Line, 100))
	}
}

func startTailing(tailer *tailer.Tailer, ruleManager *rules.Manager) {
	fmt.Println("\n=== Gerçek Zamanlı İzleme ===")
	
	logFiles := ruleManager.GetEnabledLogFiles()
	if len(logFiles) == 0 {
		fmt.Println("Etkin log dosyası bulunamadı!")
		return
	}

	fmt.Println("\nEtkin log dosyaları:")
	for i, file := range logFiles {
		fmt.Printf("%d. %s (%s)\n", i+1, file.Path, file.Type)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\nİzlenecek dosyaları seçin (virgülle ayırın, 'all' için tümünü seçin): ")
	scanner.Scan()
	selection := strings.TrimSpace(scanner.Text())

	var filesToWatch []string
	if strings.ToLower(selection) == "all" {
		for _, file := range logFiles {
			filesToWatch = append(filesToWatch, file.Path)
		}
	} else {
		indices := strings.Split(selection, ",")
		for _, idxStr := range indices {
			var idx int
			if _, err := fmt.Sscanf(strings.TrimSpace(idxStr), "%d", &idx); err == nil && idx > 0 && idx <= len(logFiles) {
				filesToWatch = append(filesToWatch, logFiles[idx-1].Path)
			}
		}
	}

	if len(filesToWatch) == 0 {
		fmt.Println("Geçerli dosya seçilmedi!")
		return
	}

	for _, filePath := range filesToWatch {
		if err := tailer.StartWatching(filePath); err != nil {
			fmt.Printf("Uyarı: %s dosyası izlenemiyor: %v\n", filePath, err)
		} else {
			fmt.Printf("✓ %s izleniyor...\n", filePath)
		}
	}

	fmt.Println("\nGerçek zamanlı izleme başlatıldı. Uyarılar aşağıda görüntülenecek.")
	fmt.Println("Durdurmak için 'q' tuşuna basın.\n")
	go func() {
		for alert := range tailer.Alerts() {
			fmt.Printf("\n[%s] %s - %s\n", alert.Severity, alert.Timestamp.Format("2006-01-02 15:04:05"), strings.Join(alert.MatchedRules, ", "))
			fmt.Printf("  Dosya: %s\n", alert.Source)
			fmt.Printf("  Satır: %s\n", truncate(alert.Line, 150))
		}
	}()
	scanner.Scan()
	if strings.ToLower(strings.TrimSpace(scanner.Text())) == "q" {
		for _, filePath := range filesToWatch {
			tailer.StopWatching(filePath)
		}
		fmt.Println("İzleme durduruldu.")
	}
}

func viewRules(ruleManager *rules.Manager) {
	fmt.Println("\n=== Kurallar ===")
	rules := ruleManager.GetRules()
	
	if len(rules) == 0 {
		fmt.Println("Kural bulunamadı!")
		return
	}

	for i, rule := range rules {
		status := "Pasif"
		if rule.Enabled {
			status = "Aktif"
		}
		fmt.Printf("\n%d. %s [%s] - %s\n", i+1, rule.Name, status, rule.Severity)
		fmt.Printf("   Desen: %s\n", rule.Pattern)
		fmt.Printf("   Açıklama: %s\n", rule.Description)
	}
}

func viewLogFiles(ruleManager *rules.Manager) {
	fmt.Println("\n=== Log Dosyaları ===")
	logFiles := ruleManager.GetLogFiles()
	
	if len(logFiles) == 0 {
		fmt.Println("Log dosyası bulunamadı!")
		return
	}

	for i, file := range logFiles {
		status := "Pasif"
		if file.Enabled {
			status = "Aktif"
		}
		
		exists := "✓"
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			exists = "✗"
		}
		
		fmt.Printf("\n%d. %s [%s] - %s\n", i+1, file.Path, status, file.Type)
		fmt.Printf("   Durum: %s\n", exists)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
