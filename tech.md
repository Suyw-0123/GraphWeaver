這是一份針對 **GraphWeaver** 專案的技術架構報告（Architecture Report）。

---

# 🛠️ GraphWeaver 技術架構報告 (v1.0)

**專案定義**：基於雲原生架構的高性能 Graph RAG 知識引擎。
**核心目標**：透過「實體-關係」建模與「語義向量」雙引擎，解決傳統 RAG 無法處理複雜關聯查詢（Multi-hop reasoning）的問題。

---

Phase 0 (Infrastructure Setup) - 本準備清單
Phase 1 (Database Schema) - 設計 MySQL Schema + 編寫 Migration
Phase 2 (Repository Layer) - 實作資料存取介面（遵循 SOLID-D）
Phase 3 (Core Service) - 實作 Ingestion Service（手動建圖）
Phase 4 (API Gateway) - 暴露 RESTful API

## 1. 核心技術棧 (Technology Stack)

| 維度 | 技術選型 | 選擇理由 (Engineering Logic) |
| :--- | :--- | :--- |
| **前端框架** | **React + TypeScript + Vite** | 現代化 Web 開發標準，組件化架構，構建速度快。 |
| **UI 樣式** | **Tailwind CSS** | Utility-first CSS 框架，快速構建響應式介面，易於維護。 |
| **後端語言** | **Go (Golang)** | 高性能併發、強型別、編譯速度快，符合 Unix 哲學。 |
| **運維平台** | **Kubernetes (kind)** | 自癒、自動擴容、基礎設施即代碼 (IaC)。 |
| **關係型數據** | **PostgreSQL** | 替代 MySQL。存儲用戶、筆記元數據、系統審計日誌。支援 JSONB 與強大的關聯查詢。 |
| **圖數據庫** | **Neo4j** | 核心 Graph RAG 引擎，處理知識點之間的拓撲關聯。 |
| **向量數據庫** | **Qdrant / Milvus** | 提供語義搜索，作為圖譜檢索的「入口」與補充。 |
| **通訊協議** | **REST / gRPC** | 對外提供標準 API，對內服務間使用高效能 gRPC。 |

---

## 2. 混合檢索架構 (The Hybrid RAG Engine)

為什麼 Graph RAG 仍需要向量資料庫？我們採用 **"Graph-Vector Hybrid"** 策略：

1.  **Vector Engine (語義模糊搜索)**：負責「找到相關的點」。當用戶問問題時，先用向量檢索找到最接近的實體（Entities）。
2.  **Graph Engine (邏輯關聯推理)**：負責「沿著線索爬行」。從向量引擎找到的點出發，在圖譜中進行 2-hop 或 3-hop 查詢，提取出隱藏的關聯邏輯。

---

## 3. 系統組件圖 (System Components)

```mermaid
graph TB
    subgraph Client_Layer [Client Layer]
        Web[Web SPA (React + Tailwind)]
    end

    subgraph K8s_Cluster [K8s Cluster: GraphWeaver-Core]
        direction TB
        
        Ingress[Ingress Controller] --> Gateway[API Gateway Service]

        subgraph Microservices [Go Microservices]
            Gateway --> Ingest[Ingestion Service]
            Gateway --> Chat[Chat/Reasoning Service]
            
            Ingest -->|Entity Extraction| LLM[LLM API / Local Model]
        end

        subgraph Persistence_Layer [Persistence Layer]
            Ingest -->|Store Metadata| Postgres[(PostgreSQL)]
            Ingest -->|Create Nodes/Edges| Neo4j[(Neo4j)]
            Ingest -->|Upsert Vectors| VectorDB[(Qdrant)]
            
            Chat --> Q_Neo4j[Cypher Query]
            Chat --> Q_Vector[Vector Search]
        end
    end

    classDef service fill:#e1f5fe,stroke:#01579b;
    classDef db fill:#fff3e0,stroke:#ff6f00;
    class Ingest,Chat,Gateway service;
    class Postgres,Neo4j,VectorDB db;
```

---

## 4. 資料流 (Data Pipeline)

### A. 數據攝取 (Ingestion Flow)
1.  **Upload**: 用戶上傳 PDF/Markdown。
2.  **Extract**: `Ingestion Service` 呼叫 LLM 提取「實體 (Entities)」與「關係 (Relations)」。
3.  **Sync**: 
    - 實體描述轉換成向量存入 **Vector DB**。
    - 實體與關係存入 **Neo4j** (例如: `A` 是 `B` 的原因)。
    - 文件原始資訊存入 **PostgreSQL**。

### B. 查詢推理 (Query Flow)
1.  **Question**: 用戶提問。
2.  **Retrieval**:
    - **向量搜索**：找出與問題語義最接近的 5 個知識點。
    - **圖譜擴散**：以這 5 個點為起點，查詢它們在 Neo4j 中 2 層以內的鄰居。
3.  **Synthesis**: 將提取到的「關聯子圖」與「原始文本」交給 LLM 生成答案。

---

## 5. 工程品質保證 (Engineering Standards)

本專案嚴格遵守 `ENGINEERING.md` 定義之規範：

-   **SOLID 原則**：
    -   `Database` 邏輯必須透過 `Interface` 封裝，以便在不同資料庫間切換（例如從 Neo4j 換成 Memgraph）。
    -   `Service` 職責單一：`Ingestion` 只管解析，`Chat` 只管推理。
-   **K8s 運維實踐**：
    -   所有配置使用 **ConfigMap** 與 **Secrets** 管理，嚴禁硬編碼。
    -   服務具備 **Health Checks (Liveness/Readiness)**，確保 K8s 能正確監控。
-   **Go 開發規範**：
    -   錯誤處理：`if err != nil` 必須明確處理，禁止忽略。
    -   並發安全：使用 `Context` 控制超時，避免 Goroutine 洩漏。

---

## 6. MVP 階段目標 (Phase 1)

1.  **基礎設施自動化**：使用 **Helm** 在 `kind` 中一鍵啟動 PostgreSQL 與 Neo4j。
2.  **前端架構搭建**：初始化 React + Vite + Tailwind 專案結構。
3.  **核心 API**：完成一個 Go 服務，能接收一段文字，手動將其轉為圖節點。
4.  **CI 整合**：每次提交代碼自動在 GitHub Action 執行測試與 Lint。

---

**報告結論**：
`GraphWeaver` 不僅僅是一個 RAG 應用，它是一個**雲原生分散式系統**。透過 K8s 的調度與 Go 的高效能，我們能構建出比單體應用（Monolith）更具韌性與擴展性的知識引擎。

---

